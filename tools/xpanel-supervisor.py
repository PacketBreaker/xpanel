#!/usr/bin/env python3
"""Xpanel managed supervisor and trusted-channel preview. Never runs upstream scripts."""
import argparse,fcntl,json,os,pathlib,signal,subprocess,sys,time,urllib.parse,urllib.request

def write(path,obj):
 temp=path.with_name(path.name+'.tmp-'+str(os.getpid()))
 fd=os.open(temp,os.O_WRONLY|os.O_CREAT|os.O_TRUNC,0o600)
 with os.fdopen(fd,'w')as f:json.dump(obj,f);f.flush();os.fsync(f.fileno())
 os.replace(temp,path)

def arguments(root):
 config=root/'CHANNEL-CONFIG.json'
 if config.is_symlink()or config.stat().st_uid!=os.getuid()or config.stat().st_mode&0o022:raise ValueError('Channel config must not be group/world writable')
 j=json.loads(config.read_text())
 if not isinstance(j,dict):raise ValueError('Channel config must be an object')
 args=['python3',str(root/'host-tools/xpanel-channel.py'),'--root',str(root)]
 for key,flag in [('manifest_url','--manifest-url'),('signature_url','--signature-url'),('trusted_key','--trusted-key')]:
  if not isinstance(j.get(key),str):raise ValueError('Missing operator channel configuration')
  args.extend([flag,j[key]])
 if not(root/'CHANNEL.json').exists():
  floor=j.get('installed_sequence')
  if type(floor)is not int or floor<1:raise ValueError('Operator installed_sequence required for first channel use')
  args.extend(['--initialize-floor',str(floor)])
 if j.get('test_loopback_http')is True:args.append('--test-loopback-http')
 return args

def command(args,timeout):
 # Own the entire subprocess group, including nested installer children.
 child=subprocess.Popen(args,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True,start_new_session=True)
 try:out,err=child.communicate(timeout=timeout)
 except BaseException:
  try:os.killpg(child.pid,signal.SIGKILL)
  except ProcessLookupError:pass
  child.communicate()
  raise
 return subprocess.CompletedProcess(args,child.returncode,out,err)

def preview(root):
 p=command(arguments(root)+['--allow-current-preview'],50)
 if p.returncode:raise ValueError('Trusted channel verification refused')
 return json.loads(p.stdout)

def stop(child):
 if child.poll()is None:
  os.killpg(child.pid,signal.SIGTERM)
  try:child.wait(timeout=30)
  except subprocess.TimeoutExpired:
   os.killpg(child.pid,signal.SIGKILL);child.wait(timeout=5)

class NoRedirect(urllib.request.HTTPRedirectHandler):
 def redirect_request(self,*a,**kw):return None

def health(root,child):
 # Operator config, loopback-only URL; no guessed public endpoint or credentials.
 j=json.loads((root/'SUPERVISOR-CONFIG.json').read_text());url=j['health_url'];u=urllib.parse.urlsplit(url)
 if u.scheme!='http' or u.hostname not in ('127.0.0.1','::1')or u.username or u.password:raise ValueError('health_url must be a literal loopback HTTP URL')
 deadline=time.monotonic()+45
 while time.monotonic()<deadline:
  if child.poll()is not None:return False
  try:
   with urllib.request.build_opener(urllib.request.ProxyHandler({}),NoRedirect()).open(url,timeout=1)as r:
    if r.status==200:return True
  except OSError:pass
  time.sleep(.25)
 return False

def recover(root,job):
 before=job.get('before')
 if before and os.readlink(root/'current')!=before:
  receipts=[f for f in(root/'history').glob('*.json')if json.loads(f.read_text()).get('to')==os.readlink(root/'current')and json.loads(f.read_text()).get('from')==before]
  if len(receipts)!=1:raise ValueError('Cannot identify rollback receipt; operator recovery required')
  result=command(['python3',str(root/'host-tools/xpanel-local.py'),'rollback','--root',str(root),'--receipt',str(receipts[0])],40)
  if result.returncode:raise ValueError('Rollback failed; operator recovery required')
 if before:
  if job.get('old_channel')is None:(root/'CHANNEL.json').unlink(missing_ok=True)
  else:write(root/'CHANNEL.json',job['old_channel'])

def serve(root):
 lock=open(root/'.supervisor.lock','a+',opener=lambda p,f:os.open(p,f,0o600));fcntl.flock(lock,fcntl.LOCK_EX|fcntl.LOCK_NB)
 terminating=False
 def signal_stop(*a):
  nonlocal terminating
  terminating=True
 signal.signal(signal.SIGTERM,signal_stop);signal.signal(signal.SIGINT,signal_stop)
 def launch():return subprocess.Popen([str(root/'current/start-managed.sh')],env={**os.environ,'XPANEL_MANAGED_ROOT':str(root)},start_new_session=True)
 status=root/'state/update-status.json';request=root/'update-request.json'
 active=root/'update-active.json'
 if active.exists():
  try:
   interrupted=json.loads(active.read_text());recover(root,interrupted);write(status,{'runId':interrupted.get('runId',''),'state':'failed','exitCode':1,'finishedAt':int(time.time())})
  except Exception:raise
  else:active.unlink()
 child=launch()
 try:
  while not terminating:
   write(root/'supervisor.json',{'at':int(time.time()),'pid':os.getpid(),'child':child.pid})
   if request.exists():
    requestlock=open(root/'.update-request.lock','a+',opener=lambda p,f:os.open(p,f,0o600));fcntl.flock(requestlock,fcntl.LOCK_EX)
    job=json.loads(request.read_text());run=job.get('runId')
    if not isinstance(run,str)or not run.isdigit():raise ValueError('Invalid queued update')
    active=root/'update-active.json';os.replace(request,active)
    write(status,{'runId':run,'state':'pending'})
    handled=False
    try:
     checked=preview(root)
     if checked['manifest_sha256']!=job.get('manifest_sha256'):raise ValueError('Channel changed after confirmation')
     old_channel=json.loads((root/'CHANNEL.json').read_text())if(root/'CHANNEL.json').exists()else None
     before=os.readlink(root/'current')
     job.update(before=before,old_channel=old_channel);write(active,job)
     stop(child)
     apply_command=arguments(root)+['--expected-manifest-sha256',checked['manifest_sha256'],'--apply']
     # Each network operation has a socket timeout, and this bounds the whole job.
     p=command(apply_command,240)
     child=launch();ready=health(root,child)
     success=p.returncode==0 and ready
     if not ready and os.readlink(root/'current')!=before:
      stop(child)
      recover(root,job)
      child=launch();health(root,child)
     write(status,{'runId':run,'state':'success'if success else'failed','exitCode':p.returncode if p.returncode else(0 if ready else 1),'finishedAt':int(time.time())})
    except (OSError,ValueError,subprocess.SubprocessError,KeyError):
     stop(child)
     recover(root,job)
     child=launch()
     write(status,{'runId':run,'state':'failed','exitCode':1,'finishedAt':int(time.time())})
     handled=True
    else:handled=True
    finally:
     if handled:active.unlink(missing_ok=True)
     requestlock.close()
   if child.poll()is not None:
    time.sleep(1)
    if not terminating:child=launch()
   time.sleep(.5)
 finally:
  stop(child);(root/'supervisor.json').unlink(missing_ok=True)

def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('command',choices=['run','preview']);p.add_argument('--root',required=True);a=p.parse_args();root=pathlib.Path(a.root).absolute()
 if root.is_symlink()or root.stat().st_uid!=os.getuid()or root.stat().st_mode&0o022 or not(root/'INSTALL.json').is_file():raise ValueError('Owned installation required')
 if a.command=='preview':print(json.dumps(preview(root)))
 else:serve(root)
if __name__=='__main__':
 try:main()
 except (OSError,ValueError,KeyError,subprocess.SubprocessError)as e:raise SystemExit('Xpanel supervisor refused: '+str(e))
