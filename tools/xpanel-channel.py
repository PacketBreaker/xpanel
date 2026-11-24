#!/usr/bin/env python3
"""Explicit, stopped-state Xpanel channel update. No built-in network destination."""
import argparse, fcntl, hashlib, ipaddress, json, os, pathlib, subprocess, tempfile, time, urllib.parse, urllib.request
MAX_MANIFEST=65536
MAX_ARCHIVE=512*1024*1024

def endpoint(url, test=False):
    u=urllib.parse.urlsplit(url)
    if u.username or u.password or u.fragment: raise ValueError('Credentials/fragments are forbidden in channel URLs')
    if u.scheme=='https' and u.hostname:return url
    if test and u.scheme=='http' and u.hostname:
        try:
            if ipaddress.ip_address(u.hostname).is_loopback:return url
        except ValueError:pass
    raise ValueError('HTTPS required; only literal loopback HTTP allowed in explicit test mode')

class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self,*a,**kw):raise ValueError('Redirect refused; channel must provide final URLs')

def download(url, path, limit, test=False):
    endpoint(url,test)
    opener=urllib.request.build_opener(NoRedirect())
    with opener.open(url,timeout=20) as response,open(path,'xb')as out:
        size=0
        while True:
            b=response.read(min(1024*1024,limit-size+1))
            if not b:break
            size+=len(b)
            if size>limit:raise ValueError('Download exceeds size limit')
            out.write(b)
    os.chmod(path,0o600)
    return size

def verify(manifest, signature, key):
    # Signature covers the exact downloaded bytes, never a reserialized object.
    subprocess.run(['openssl','pkeyutl','-verify','-pubin','-inkey',str(key),'-rawin','-in',str(manifest),'-sigfile',str(signature)],check=True,capture_output=True)

def main():
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--allow-current-preview',action='store_true');p.add_argument('--expected-manifest-sha256');p.add_argument('--manifest-url',required=True);p.add_argument('--signature-url',required=True);p.add_argument('--trusted-key',required=True)
    p.add_argument('--initialize-floor',type=int,help='Explicit operator-confirmed installed release sequence for first channel use');p.add_argument('--root',required=True);p.add_argument('--apply',action='store_true');p.add_argument('--test-loopback-http',action='store_true')
    a=p.parse_args();root=pathlib.Path(a.root).absolute();key=pathlib.Path(a.trusted_key).absolute()
    if not key.is_file():raise ValueError('Out-of-band trusted Ed25519 public key required')
    if root.is_symlink() or not root.is_dir():raise ValueError('Existing non-symlink owned root required')
    lock=open(root/'.xpanel-channel.lock','a+',opener=lambda p,f:os.open(p,f,0o600))
    try:fcntl.flock(lock,fcntl.LOCK_EX|fcntl.LOCK_NB)
    except BlockingIOError:raise ValueError('Another channel operation is active')
    # Resolve no keys, endpoints, permissions or production identity from the feed.
    with tempfile.TemporaryDirectory(prefix='xpanel-channel-')as tmp:
        d=pathlib.Path(tmp);os.chmod(d,0o700);m=d/'release.json';sig=d/'release.sig'
        download(a.manifest_url,m,MAX_MANIFEST,a.test_loopback_http);download(a.signature_url,sig,1024,a.test_loopback_http);verify(m,sig,key)
        manifest_sha=hashlib.sha256(m.read_bytes()).hexdigest()
        if a.expected_manifest_sha256 and manifest_sha!=a.expected_manifest_sha256:raise ValueError('Channel changed after browser confirmation; check again')
        j=json.loads(m.read_text())
        if not isinstance(j,dict):raise ValueError('Manifest must be a JSON object')
        if j.get('schema')!=1 or j.get('product')!='Xpanel' or j.get('architecture')!='linux-amd64':raise ValueError('Wrong channel product/schema/architecture')
        seq=j.get('sequence');expires=j.get('expires_unix');size=j.get('bytes');sha=j.get('sha256','')
        if type(seq)is not int or seq<1 or type(expires)is not int or expires<=int(time.time()):raise ValueError('Expired or invalid release sequence')
        if type(size)is not int or not 0<size<=MAX_ARCHIVE or len(sha)!=64 or any(c not in '0123456789abcdef'for c in sha):raise ValueError('Invalid package size/hash')
        endpoint(j['archive_url'],a.test_loopback_http)
        state=root/'CHANNEL.json';previous=json.loads(state.read_text())if state.exists()else{}
        if not isinstance(previous,dict):raise ValueError('Invalid local channel receipt')
        if not state.exists():
            if a.initialize_floor is None or a.initialize_floor<1:raise ValueError('First use requires explicit --initialize-floor with the installed release sequence; obtain it from the installed release publisher, never infer it from this feed')
            previous={'sequence':a.initialize_floor}
        elif a.initialize_floor is not None:raise ValueError('Initialization floor is only allowed before a channel receipt exists')
        if seq<previous.get('sequence',0) or (seq==previous.get('sequence',0) and (a.apply or not a.allow_current_preview)):raise ValueError('Channel rollback/replay refused; use explicit local receipt rollback instead')
        if not(root/'INSTALL.json').is_file():raise ValueError('Use local installer to create an owned installation first')
        if not a.apply:
            print(json.dumps({'verified':True,'update_available':seq>previous.get('sequence',0),'manifest_sha256':manifest_sha,'sequence':seq,'bytes':size,'sha256':sha,'apply':False,'note':'No package downloaded or installation changed. Explicit --apply performs a stopped-state update.'}));return
        archive=d/'package.zip';actual=download(j['archive_url'],archive,MAX_ARCHIVE,a.test_loopback_http)
        h=hashlib.sha256()
        with open(archive,'rb')as f:
            for block in iter(lambda:f.read(1024*1024),b''):h.update(block)
        if actual!=size or h.hexdigest()!=sha:raise ValueError('Package bytes/hash mismatch')
        # The existing installer owns lifetime locking, product inventory, DB backup,
        # release switch and explicit rollback. Never start a process or service.
        installer=pathlib.Path(__file__).with_name('xpanel-local.py')
        subprocess.run(['python3',str(installer),'update','--root',str(root),'--archive',str(archive),'--sha256',sha],check=True)
        receipt={'sequence':seq,'sha256':sha,'manifest_sha256':hashlib.sha256(m.read_bytes()).hexdigest(),'trusted_key_sha256':hashlib.sha256(key.read_bytes()).hexdigest()}
        temp=root/('.channel-'+str(time.time_ns())+'.json')
        fd=os.open(temp,os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o600)
        with os.fdopen(fd,'w')as f:json.dump(receipt,f);f.flush();os.fsync(f.fileno())
        os.replace(temp,state)
        print(json.dumps({'updated':True,'sequence':seq,'note':'No process started. Verify and start the managed release explicitly.'}))
if __name__=='__main__':
    try:main()
    except (OSError,ValueError,KeyError,subprocess.CalledProcessError)as e:raise SystemExit('Xpanel channel update refused: '+str(e))
