#!/usr/bin/env python3
"""Xpanel owned-directory installer. No network, service, or root access needed."""
import argparse, contextlib, fcntl, hashlib, json, os, pathlib, shutil, sqlite3, stat, tempfile, time, zipfile

MARKER = 'xpanel-owned-local-install-v2'

def digest(path):
    h=hashlib.sha256()
    with open(path,'rb') as f:
        for block in iter(lambda:f.read(1024*1024),b''): h.update(block)
    return h.hexdigest()

def private_json(path, value):
    with open(path,'x',opener=lambda p,f:os.open(p,f,0o600)) as f:
        json.dump(value,f,indent=2); f.flush(); os.fsync(f.fileno())

def validate_extract(archive, expected, dest):
    # Capture one private snapshot before checking or extracting. The caller may
    # replace/change its file, but only this hashed snapshot is ever extracted.
    with tempfile.TemporaryFile() as snapshot:
        h=hashlib.sha256()
        with open(archive,'rb') as source:
            for block in iter(lambda:source.read(1024*1024),b''):
                h.update(block);snapshot.write(block)
        if len(expected)!=64 or h.hexdigest()!=expected.lower():raise ValueError('Archive SHA256 mismatch; nothing installed')
        snapshot.seek(0)
        return extract_verified_snapshot(snapshot,dest)

def extract_verified_snapshot(snapshot,dest):
    with zipfile.ZipFile(snapshot) as z:
        infos=z.infolist()
        if len(infos)>20000 or sum(i.file_size for i in infos)>2_000_000_000: raise ValueError('Archive exceeds local package limits')
        roots=set(); names=set()
        for i in infos:
            p=pathlib.PurePosixPath(i.filename)
            if p.is_absolute() or '..' in p.parts or '\\' in i.filename or not p.parts: raise ValueError('Unsafe archive path')
            if i.filename in names: raise ValueError('Duplicate archive entry')
            names.add(i.filename); roots.add(p.parts[0])
            mode=i.external_attr>>16
            if stat.S_ISLNK(mode) or (stat.S_IFMT(mode) not in (0,stat.S_IFREG,stat.S_IFDIR)): raise ValueError('Special archive entry refused')
            if len(p.parts)>1 and p.parts[1] in ('state','logs'): raise ValueError('Runtime archive contains private state or logs')
        if len(roots)!=1: raise ValueError('Expected one package root')
        root=next(iter(roots)); prefix=root+'/'
        for i in infos:
            if i.is_dir(): continue
            rel=pathlib.PurePosixPath(i.filename).relative_to(root)
            if not rel.parts: raise ValueError('Expected package directory')
            out=dest/pathlib.Path(*rel.parts);out.parent.mkdir(parents=True,exist_ok=True,mode=0o700)
            with z.open(i) as src, open(out,'xb') as dst: shutil.copyfileobj(src,dst)
            os.chmod(out,0o700 if i.external_attr>>16 & 0o111 else 0o600)
    manifest={}
    for line in (dest/'SHA256SUMS').read_text().splitlines():
        h,rel=line.split('  ',1);p=pathlib.PurePosixPath(rel)
        if p.is_absolute() or '..' in p.parts or rel in manifest: raise ValueError('Unsafe or duplicate manifest entry')
        manifest[rel]=h
    actual={str(p.relative_to(dest)) for p in dest.rglob('*') if p.is_file()}
    if actual != set(manifest)|{'SHA256SUMS'}: raise ValueError('Package file inventory mismatch')
    for rel,h in manifest.items():
        if digest(dest/rel)!=h: raise ValueError('File hash mismatch: '+rel)
    for required in ('xpanel','xpanel-bootstrap','start-local.sh','start-managed.sh','PACKAGE-PROVENANCE.json','licenses/3x-ui-GPLv3.txt'):
        if required not in actual: raise ValueError('Missing package file: '+required)
    provenance=json.loads((dest/'PACKAGE-PROVENANCE.json').read_text())
    if provenance.get('architecture')!='linux-amd64': raise ValueError('Only linux-amd64 supported')
    return provenance

@contextlib.contextmanager
def locked(root):
    state=root/'state'; state.mkdir(mode=0o700,exist_ok=True)
    if state.is_symlink(): raise ValueError('Shared state cannot be a symlink')
    with open(state/'.xpanel-runtime.lock','a+',opener=lambda p,f:os.open(p,f,0o600)) as lock:
        try: fcntl.flock(lock,fcntl.LOCK_EX|fcntl.LOCK_NB)
        except BlockingIOError: raise ValueError('Xpanel is running or another operation owns this install; stop it first')
        yield

def check_root(root, create=False):
    if root.is_symlink(): raise ValueError('Install root cannot be a symlink')
    if not root.exists():
        if not create: raise ValueError('Owned installation not found')
        root.mkdir(mode=0o700,parents=False)
        private_json(root/'INSTALL.json',{'kind':MARKER})
    if root.stat().st_uid!=os.getuid(): raise ValueError('Install root belongs to another user')
    if not (root/'INSTALL.json').is_file():raise ValueError('Refusing foreign directory without Xpanel ownership marker')
    meta=json.loads((root/'INSTALL.json').read_text())
    if meta.get('kind')!=MARKER: raise ValueError('Refusing incompatible managed layout or foreign directory; v1 roots require a new v2 install, no automatic migration')
    for folder in ('releases','history','state','logs','runtime-bin'):
        p=root/folder
        if p.is_symlink(): raise ValueError('Owned directory cannot be a symlink: '+folder)
        p.mkdir(mode=0o700,exist_ok=True)
    return meta

def current(root):
    link=root/'current'
    if not link.exists() and not link.is_symlink(): return None
    if not link.is_symlink(): raise ValueError('Current release must be an owned relative symlink')
    target=os.readlink(link);p=pathlib.PurePosixPath(target)
    if len(p.parts)!=2 or p.parts[0]!='releases' or '..' in p.parts: raise ValueError('Current release escapes owned releases')
    if not (root/target).is_dir(): raise ValueError('Current release missing')
    return target

def switch(root,target):
    tmp=root/('.current-'+str(time.time_ns()));os.symlink(target,tmp);os.replace(tmp,root/'current')

def backup_db(root,label):
    db=root/'state/x-ui.db'
    if not db.exists():return None
    if db.is_symlink():raise ValueError('Database cannot be a symlink')
    out=root/'history'/(label+'.db')
    with sqlite3.connect(db.as_uri()+'?mode=ro',uri=True) as src, sqlite3.connect(out) as dst: src.backup(dst)
    os.chmod(out,0o600);return out.name

def apply(args):
    root=pathlib.Path(args.root).absolute();check_root(root,create=args.command=='install')
    with locked(root):
        previous=current(root)
        if args.command=='install' and previous:raise ValueError('Already installed; use update')
        if args.command=='update' and not previous:raise ValueError('No current release to update')
        stage=pathlib.Path(tempfile.mkdtemp(prefix='.stage-',dir=root/'releases'))
        try:
            provenance=validate_extract(args.archive,args.sha256,stage)
            label=str(time.time_ns());snapshot=backup_db(root,label)
            target='releases/'+label
            # Core config/geodata updates go to a separate private copy. Never
            # let a managed launch modify the verified release tree.
            runtime=root/'runtime-bin'/label
            shutil.copytree(stage/'bin',runtime)
            for f in runtime.rglob('*'):
                if f.is_file():os.chmod(f,0o700 if f.stat().st_mode & 0o111 else 0o600)
            os.rename(stage,root/target)
            receipt={'from':previous,'to':target,'snapshot':snapshot,'archive_sha256':args.sha256,'provenance':provenance}
            private_json(root/'history'/(label+'.json'),receipt)
            switch(root,target)
            print(json.dumps({'current':str(root/'current'),'start':str(root/'current/start-managed.sh'),'previous':previous,'database_backup':snapshot,'receipt':str(root/'history'/(label+'.json')),'note':'No process or service started. Use start-managed.sh, not the unmanaged launcher.'},indent=2))
        finally:
            if stage.exists():shutil.rmtree(stage)

def rollback(args):
    root=pathlib.Path(args.root).absolute();check_root(root)
    with locked(root):
        receipt_path=pathlib.Path(args.receipt).absolute()
        if receipt_path.parent!=root/'history':raise ValueError('Receipt must be in this install history')
        j=json.loads(receipt_path.read_text());now=current(root)
        if now!=j['to'] or not j['from']:raise ValueError('Receipt is not for current update, or no previous release exists')
        for target in (j['from'],j['to']):
            p=pathlib.PurePosixPath(target)
            if len(p.parts)!=2 or p.parts[0]!='releases' or '..' in p.parts:raise ValueError('Unsafe receipt release')
        if not (root/j['from']).is_dir():raise ValueError('Previous release missing')
        snapshot=j['snapshot']
        if not snapshot and (root/'state/x-ui.db').exists():raise ValueError('Previous release had no database snapshot; refusing to pair an older release with newly created state')
        if snapshot and pathlib.PurePath(snapshot).name!=snapshot:raise ValueError('Unsafe snapshot path')
        label='before-rollback-'+str(time.time_ns());saved=backup_db(root,label)
        if snapshot:
            src=root/'history'/snapshot
            if not src.is_file() or src.is_symlink():raise ValueError('Prior database snapshot missing')
            tmp=root/'state/x-ui.db.rollback-temp';shutil.copyfile(src,tmp);os.chmod(tmp,0o600)
            # No managed process can hold a connection while this lock is held.
            for suffix in ('-wal','-shm'):
                p=root/'state'/('x-ui.db'+suffix)
                if p.exists():p.unlink()
            os.replace(tmp,root/'state/x-ui.db')
        switch(root,j['from'])
        print(json.dumps({'current':j['from'],'restored_snapshot':snapshot,'before_rollback_snapshot':saved,'note':'Stopped-state rollback completed; start-managed.sh and verify before reconnecting clients.'},indent=2))

def main():
    p=argparse.ArgumentParser(description=__doc__);sub=p.add_subparsers(dest='command',required=True)
    for command in ('install','update'):
        s=sub.add_parser(command);s.add_argument('--root',required=True);s.add_argument('--archive',required=True);s.add_argument('--sha256',required=True)
    s=sub.add_parser('rollback');s.add_argument('--root',required=True);s.add_argument('--receipt',required=True)
    args=p.parse_args()
    try:
        if args.command=='rollback':rollback(args)
        else:apply(args)
    except (OSError,ValueError,KeyError,sqlite3.Error,zipfile.BadZipFile) as e:p.exit(1,'Xpanel local operation refused: '+str(e)+'\n')
if __name__=='__main__':main()
