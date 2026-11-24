# Xpanel portable development package

This is a separate branded derivative with the original upstream license and credit intact. It does not install a system service or modify an existing installation. Run start-local.sh from a private directory. First-admin setup requires a local terminal with hidden password input. The panel initially listens only on 127.0.0.1:2053, and subscriptions are disabled until configured. Accessing this from another machine requires a separately approved deployment/network plan.

Never copy an existing 3x-ui database into state. The launcher refuses unmarked databases rather than silently migrating/adopting them. Backups and migrations need an explicit supported workflow. Do not start a second instance on the same state directory.

The original upstream update/install/service scripts are preserved under upstream-scripts for attribution and comparison, not as Xpanel installation tools. Top-level entries refuse to run. Automatic update is unavailable until a branded verified update channel exists. Do not run scripts from upstream to update this derivative: they replace Xpanel with a different product. No full feature or production security claim is made by this development package.

First run must be in a local terminal. It asks for a username and a hidden password twice. No initial password is printed or put into ordinary logs. Once an admin exists, later launches can be unattended with the same private state directory. Do not run two processes against one state directory.

For local recovery, stop Xpanel first, keep its database backed up privately, then from the package directory run:

    XUI_DB_FOLDER="$PWD/state" ./xpanel recover-admin

This requires a local controlling terminal and hidden password confirmation. It supports exactly one existing admin, preserves its ID and data, and increments its login epoch so old sessions become invalid. It does not change two-factor configuration. An empty or multi-admin database is refused. Anyone who can write these local state files already controls the installation; keep the directory private. The original setting -username/-password route is retained upstream compatibility, but do not use it for recovery because a command-line password can enter shell history or process listings. Never send a password in chat.
