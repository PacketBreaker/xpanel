#!/usr/bin/env python3
"""Prepare a non-root Xpanel systemd service. System installation is explicit."""
import argparse,json,os,pathlib,re,shutil,subprocess,sys,urllib.parse,pwd,grp,stat,fcntl,sqlite3,hashlib

def write_new(path,text):
 fd=os.open(path,os.O_WRONLY|os.O_CREAT|os.O_EXCL|os.O_NOFOLLOW,0o644)
 with os.fdopen(fd,'w')as f:
  os.fchmod(f.fileno(),0o644);f.write(text)

def prepare(root,health):
 # This function never runs privileged. The service owner may modify its own
 # files, but cannot use preparation to make root follow/chown its paths.
 if os.getuid()==0:raise ValueError('Privileged file preparation refused')
 if root.is_symlink()or root.stat().st_uid!=os.getuid()or root.stat().st_mode&0o022:raise ValueError('Unsafe or foreign installation root')
 if json.loads((root/'INSTALL.json').read_text()).get('kind')!='xpanel-owned-local-install-v2':raise ValueError('Wrong install product/layout')
 current=root/'current'
 if not current.is_symlink():raise ValueError('A complete managed current release is required')
 target=os.readlink(current)
 if not re.fullmatch(r'releases/[0-9]+',target):raise ValueError('Invalid managed current release')
 release=root/target
 for required in ['xpanel','xpanel-bootstrap','start-managed.sh','SHA256SUMS','PACKAGE-PROVENANCE.json']:
  p=release/required
  if not p.is_file()or p.is_symlink():raise ValueError('Incomplete managed current release')
 inventory={}
 for line in(release/'SHA256SUMS').read_text().splitlines():
  h,n=line.split('  ',1);rel=pathlib.PurePosixPath(n)
  if rel.is_absolute()or '..'in rel.parts or n in inventory:raise ValueError('Invalid current inventory')
  inventory[n]=h
 actual={str(p.relative_to(release))for p in release.rglob('*')if p.is_file()}
 if actual!=set(inventory)|{'SHA256SUMS'}:raise ValueError('Incomplete current inventory')
 for n,h in inventory.items():
  dig=hashlib.sha256()
  with open(release/n,'rb')as f:
   for block in iter(lambda:f.read(1024*1024),b''):dig.update(block)
  if dig.hexdigest()!=h:raise ValueError('Current release hash mismatch')
 if not(root/'runtime-bin'/pathlib.Path(target).name).is_dir():raise ValueError('Current runtime directory missing')
 for parent in root.parents:
  st=parent.stat()
  if parent.is_symlink()or(st.st_mode&0o002 and not st.st_mode&stat.S_ISVTX):raise ValueError('Unsafe installation ancestor')
  bit=stat.S_IXUSR if st.st_uid==os.getuid()else(stat.S_IXGRP if st.st_gid==os.getgid()else stat.S_IXOTH)
  if not st.st_mode&bit:raise ValueError('Service user cannot traverse installation ancestors')
 for entry in root.rglob('*'):
  if entry.is_symlink():
   if entry!=current:raise ValueError('Unsafe managed tree symlink')
   continue
  st=entry.stat()
  if st.st_uid!=os.getuid()or st.st_mode&0o022:raise ValueError('Managed tree must be service-owned and not group/world writable')
  if not(stat.S_ISREG(st.st_mode)or stat.S_ISDIR(st.st_mode)):raise ValueError('Special file refused in managed tree')
 rootfd=os.open(root,os.O_RDONLY|os.O_DIRECTORY|os.O_NOFOLLOW);statefd=None;locks=[]
 try:
  statefd=os.open('state',os.O_RDONLY|os.O_DIRECTORY|os.O_NOFOLLOW,dir_fd=rootfd)
  for directory,name in[(rootfd,'.supervisor.lock'),(statefd,'.xpanel-runtime.lock')]:
   fd=os.open(name,os.O_RDWR|os.O_CREAT|os.O_NOFOLLOW,0o600,dir_fd=directory);locks.append(fd)
   st=os.fstat(fd)
   if not stat.S_ISREG(st.st_mode)or st.st_uid!=os.getuid()or st.st_nlink!=1 or st.st_mode&0o022:raise ValueError('Unsafe service lock')
   fcntl.flock(fd,fcntl.LOCK_EX|fcntl.LOCK_NB)
  db=root/'state/x-ui.db'
  with sqlite3.connect('file:'+str(db)+'?mode=ro',uri=True)as c:
   if c.execute('SELECT count(*) FROM users').fetchone()[0]<1:raise ValueError('Complete local first-admin setup')
  tools=root/'host-tools';tools.mkdir(mode=0o700,exist_ok=True)
  for name in ['xpanel-supervisor.py','xpanel-channel.py','xpanel-local.py']:
   shutil.copyfile(pathlib.Path(__file__).with_name(name),tools/name);os.chmod(tools/name,0o700)
  config=root/'SUPERVISOR-CONFIG.json'
  if not config.exists():
   fd=os.open(config,os.O_WRONLY|os.O_CREAT|os.O_EXCL|os.O_NOFOLLOW,0o600)
   with os.fdopen(fd,'w')as f:json.dump({'health_url':health},f)
 finally:
  for fd in locks:os.close(fd)
  if statefd is not None:os.close(statefd)
  os.close(rootfd)

def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--root',required=True);p.add_argument('--service-user');p.add_argument('--unit-output',required=True);p.add_argument('--health-url',default='http://127.0.0.1:2053/');p.add_argument('--service-name',default='xpanel');p.add_argument('--install-system',action='store_true');a=p.parse_args()
 # Permission refusal must precede all output, lock or host-tools mutations.
 if a.install_system and os.geteuid()!=0:raise ValueError('System installation requires root; nothing prepared or changed')
 account=pwd.getpwnam(a.service_user)if a.service_user else pwd.getpwuid(os.getuid())
 if account.pw_uid==0:raise ValueError('A non-root --service-user is required')
 if os.getuid()not in(0,account.pw_uid):raise ValueError('Run as install owner, or root with explicit service user')
 group=grp.getgrgid(account.pw_gid).gr_name
 for name in[account.pw_name,group]:
  if not re.fullmatch(r'[a-zA-Z_][a-zA-Z0-9_.-]*[$]?',name):raise ValueError('Unsupported account/group name')
 if not re.fullmatch('[a-zA-Z0-9_-]+',a.service_name):raise ValueError('Invalid service name')
 root=pathlib.Path(a.root).absolute();out=pathlib.Path(a.unit_output).absolute()
 if not re.fullmatch(r'/[a-zA-Z0-9_./-]+',str(root))or '..'in root.parts:raise ValueError('Use simple absolute root')
 u=urllib.parse.urlsplit(a.health_url)
 if u.scheme!='http'or u.hostname not in('127.0.0.1','::1')or u.username or u.password:raise ValueError('health-url must use literal loopback HTTP')
 if out.exists()or out.is_symlink():raise ValueError('Existing unit output refused')
 if os.getuid()==0:
  # All user-tree validation, locks and writes occur after dropping privileges.
  # No privileged O_CREAT/fchown/read of lock paths, even before validation.
  childout=root/('.service-prepare-'+str(os.getpid())+'.service')
  def demote():os.setgroups([]);os.setgid(account.pw_gid);os.setuid(account.pw_uid)
  subprocess.run([sys.executable,str(pathlib.Path(__file__).absolute()),'--root',str(root),'--service-user',account.pw_name,'--unit-output',str(childout),'--health-url',a.health_url,'--service-name',a.service_name],check=True,preexec_fn=demote)
  # Never read or install a user-writable unit. Preview cleanup is unprivileged.
  subprocess.run([sys.executable,'-c','import os,sys;os.unlink(sys.argv[1])',str(childout)],check=True,preexec_fn=demote)
 else:prepare(root,a.health_url)
 unit='''[Unit]
Description=Xpanel managed panel and update supervisor
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User={user}
Group={group}
CapabilityBoundingSet=
AmbientCapabilities=
RestrictSUIDSGID=true
WorkingDirectory={root}
ExecStart=/usr/bin/python3 {root}/host-tools/xpanel-supervisor.py run --root {root}
Restart=on-failure
RestartSec=3
KillMode=mixed
TimeoutStopSec=45
UMask=0077
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
'''.format(root=root,user=account.pw_name,group=group)
 write_new(out,unit)
 if a.install_system:
  target=pathlib.Path('/etc/systemd/system')/(a.service_name+'.service')
  subprocess.run(['systemd-analyze','verify',str(out)],check=True)
  write_new(target,unit)
  for args in[['daemon-reload'],['enable',a.service_name+'.service'],['start',a.service_name+'.service']]:subprocess.run(['systemctl']+args,check=True)
  print(json.dumps({'installed':str(target),'service':a.service_name+'.service'}))
 else:print(json.dumps({'prepared':str(out),'service_user':account.pw_name,'uid':account.pw_uid,'capabilities':[],'note':'No system service changed or started'}))
if __name__=='__main__':
 try:main()
 except (OSError,ValueError,KeyError,sqlite3.Error,subprocess.SubprocessError)as e:raise SystemExit('Xpanel service setup refused: '+str(e))
