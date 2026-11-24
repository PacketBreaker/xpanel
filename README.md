# Xpanel

**Arabic / العربية · English**

A bilingual web panel for Xray and related proxy services, adapted from **3x-ui v3.11.0**, pinned to upstream commit `c621ba9014a6aa21fd7c9ecf45655e44e24ebf13`.

**Eng.Abdelrahman Rabie 2026**

> Development candidate v20. Local panel, proxy-traffic, update and recovery tests have been run. This is not a production-security certification or a claim that every deployment mode works on every host. No server is installed or changed by opening this repository.

## العربية

Xpanel لوحة تحكم بواجهة عربية وإنجليزية لإدارة إعدادات البروكسي والعملاء والاشتراكات والإحصاءات. مبنية على 3x-ui مع الحفاظ على ترخيص GPL وحقوق المطورين الأصليين، وإضافات Xpanel موضحة في ملفات المراجعة.

### ما الذي تم تجهيزه؟

- واجهة عربية باتجاه RTL وإنجليزية باتجاه LTR، مع إدارة العملاء والحصص والصلاحية والاشتراكات وQR والإحصاءات وإعدادات التوجيه والنودز الموروثة من المشروع الأصلي.
- تثبيت محلي داخل مجلد مملوك لحساب الخدمة، وإنشاء أول مدير تفاعليا دون بيانات دخول افتراضية.
- تحديث من المتصفح عبر قناة Xpanel موقعة يحددها المشغل، مع إيقاف اللوحة وتحديثها وإعادة تشغيلها، ونسخة SQLite للتراجع عند فشل بدء الإصدار الجديد.
- إعداد خدمة systemd بحساب غير root، دون منح قدرات شبكية خاصة تلقائيا.

وجود ميزة في السورس لا يعني أنها اختبرت في كل بيئات التشغيل. دعم SSH الذي طلب سابقا ليس ضمن نطاق Xpanel هذا.

### التثبيت

الحزمة الجاهزة مخصصة لـ **Linux amd64** وتعتمد على glibc برموز تصل إلى `GLIBC_2.34`. تحتاج Python 3 و`flock` وOpenSSL يدعم Ed25519، وsystemd فقط إذا أردت تثبيت خدمة للنظام.

1. نزّل حزمة التشغيل `xpanel-private-linux-amd64-development-2026-11-24-v20.zip` من قسم Releases. كلمة private في اسم الأرشيف تصف تسمية النسخة المحلية الأصلية، وليست صلاحية وصول المستودع.
2. تحقق من SHA256 المبين أدناه، ثم فك الحزمة في مجلد تجهيز خاص بك. لا تشغّل سكربتات تثبيت 3x-ui القديمة الموجودة ضمن السورس باعتبارها مثبت Xpanel.
3. استخدم حساب خدمة موجودا غير root. أنشئ المجلد الأب بصلاحيات مناسبة، لكن اترك مجلد التثبيت الجديد نفسه للمثبت. نفذ أوامر التثبيت أدناه بحساب الخدمة.
4. شغّل `current/start-managed.sh` من طرفية تفاعلية لإنشاء أول مدير. اللوحة تبدأ على `127.0.0.1:2053` والاشتراكات معطلة حتى تضبطها. للوصول من جهاز آخر استخدم نفقا آمنا أو إعداد reverse proxy راجعته بنفسك. لا تكشف لوحة الإدارة للإنترنت دون حماية.
5. أوقف التشغيل التفاعلي قبل إعداد الخدمة. جهّز ملف الخدمة، راجعه، ثم نفذ خطوة تثبيت النظام صراحة بصلاحيات root مع اسم حساب الخدمة نفسه.

```sh
# From the extracted runtime directory, as the non-root install owner:
python3 xpanel-local.py install \
  --root /opt/xpanel \
  --archive /absolute/path/xpanel-private-linux-amd64-development-2026-11-24-v20.zip \
  --sha256 7ceaa33e3093c25cdedc28dd5b7cc3e8fe5f245351361d5fba9dcec7a70ba0f3

/opt/xpanel/current/start-managed.sh

# Stop the interactive panel before preparing the service.
python3 xpanel-service.py --root /opt/xpanel \
  --service-user xpanel --unit-output /tmp/xpanel-preview.service
systemd-analyze verify /tmp/xpanel-preview.service

# Explicit system installation, run as root only after reviewing the unit:
python3 xpanel-service.py --root /opt/xpanel \
  --service-user xpanel --unit-output /root/xpanel-install.service --install-system
systemctl status xpanel.service
```

`xpanel` اسم مثال لحساب موجود، وليس حسابا ينشئه البرنامج. يجب أن يمتلك الحساب مجلد التثبيت وملفاته، وأن يستطيع المرور عبر المجلدات الأب. المثبت يرفض المجلدات الأجنبية أو النسخ الناقصة ولا يستورد قاعدة بيانات قديمة تلقائيا. إذا تغير المنفذ أو مسار اللوحة، حدد `--health-url` بعنوان HTTP محلي حرفي مناسب، مثل `http://127.0.0.1:2053/`.

### الاستخدام والتحديث

بعد تسجيل الدخول، اضبط inbound والعملاء والحصة وتاريخ الانتهاء، ثم تحقق من الاتصال بعميل حقيقي قبل الاعتماد على الخدمة. فعّل الاشتراكات وعناوينها فقط بعد مراجعة إعدادات الوصول. احفظ نسخا احتياطية خاصة، لأنها تحتوي على بيانات حساسة.

التحديث من المتصفح لا ينزل نسخة 3x-ui الأصلية بدلا من Xpanel. يجب أولا أن يضبط المشغل ملف `CHANNEL-CONFIG.json` ومفتاح التحقق والقناة الموقعة وتسلسل الإصدار المثبت من مصدر مستقل. لا توجد قناة إنتاج أو مفاتيح خاصة مضمنة. راجع `XPANEL-MANAGED-UPDATES-v16.md` في السورس أو `MANAGED-UPDATES.md` داخل حزمة التشغيل للأوامر وشروط الثقة والتراجع وصيانة host-tools.

### حدود مهمة

- تشغيل systemd الفعلي عند الإقلاع ومسار خفض صلاحيات عملية root لم يختبرا في بيئة المراجعة المتاحة. توليد الوحدة والتحقق الثابت والتشغيل المحلي بحساب غير root اجتازت الاختبارات.
- الخدمة لا تمنح CAP_NET_ADMIN أو CAP_NET_RAW أو CAP_NET_BIND_SERVICE. أوضاع kernel TUN/WireGuard وتعديل المسارات والجدار الناري والمنافذ المنخفضة التي تتطلب صلاحيات ليست مضمونة بهذا الإعداد.
- فحص بدء التشغيل هو استجابة HTTP200 للوحة مع بقاء العملية، وليس إثبات جاهزية كل بروتوكولات البروكسي.
- قياس الحصص دوري وقد يسمح بتجاوز قبل الإنفاذ. لا يوجد حد زمني أو حجم تجاوز مضمون.
- PostgreSQL الحي، مزودو الخدمات الخارجية، وكل حالات انقطاع الكهرباء/القتل المفاجئ ليست مؤهلة بالكامل. أدوات إدارة الخدمة المثبتة في host-tools تحدث يدويا بعد إيقاف الخدمة، ولا تستبدل نفسها تلقائيا.

## English

Xpanel provides Arabic RTL and English LTR administration for the upstream panel's client, inbound, subscription, traffic, routing and node features. It adds owned-directory installation, first-admin setup, operator-configured signed browser updates and a non-root systemd service preparation path.

The frozen **v20 runtime and source archives** are the reviewed handoff artifacts. The repository's introduction may be updated separately for documentation. Keep the original GPL license, third-party notices and adaptation credit. Do not treat retained upstream install/update scripts as the Xpanel installation path.

### Install and operate

Use the numbered Arabic section and commands above; the commands are language-independent. Install into a **new directory owned by an existing non-root account**. The parent directory must already be writable by that account. The example `/opt/xpanel` requires its parent permissions to be deliberately prepared by the operator; the installer does not create users or grant privileges.

Start `current/start-managed.sh` interactively once and create the first administrator. There are no shipped default credentials. The default panel listener is loopback-only on port2053; subscriptions start disabled. Stop it before service preparation. Review the generated unit, then explicitly install it as root with `--service-user` naming the same non-root owner. Existing system units are not overwritten. Check `systemctl status`, local login and real intended client traffic yourself.

The service has explicit User/Group, empty capability sets and NoNewPrivileges. It does not promise kernel networking, firewall changes, raw sockets or privileged low ports. No automatic in-place migration of another panel or existing database is included.

### Updates and recovery

Signed browser updates require an operator-owned `CHANNEL-CONFIG.json`, an independently delivered trusted public key, HTTPS manifest/signature endpoints, and an independently grounded installed sequence. None are configured for production in this handoff. Browser confirmation binds the exact signed manifest hash. The supervisor stops the panel, applies the verified owned package, restarts it and checks panel HTTP200/process liveness. A failed new-panel startup restores the previous release, SQLite snapshot and channel receipt. This does not establish all-protocol readiness or power-loss atomicity.

In the complete source ZIP, see `XPANEL-MANAGED-UPDATES-v16.md`, `XPANEL-LOCAL-INSTALL.md`, `XPANEL-CHANNEL-v15.md`, and `XPANEL-OPERATOR-RECOVERY.md`. The older local-installer document describes the standalone stopped-state tool; the newer managed-updates document governs the supervised browser/service workflow.

### Verification and limits

Local tests exercised actual panel login, proxy payloads for multiple protocol paths, signed browser updates, archive/hash/replay refusals, a failing new release with rollback and relogin, and non-root service-layout checks. Independent review reproduced selected core and lifecycle cases. Test coverage is not a promise of full upstream parity in every environment.

Actual systemd boot/start and root-to-service-user execution remain untested here. Live PostgreSQL, real external provider accounts, privileged networking and every interruption point remain qualification limits. Poll-based traffic enforcement may overshoot. Keep backups private and review deployment security before exposure. GitHub Actions is disabled on this repository so preserved upstream workflows do not publish or deploy automatically.

## Source access

The complete source archive is supplied alongside the full runtime in Releases. Browsable source-tree import is pending a supported authenticated bulk git route; this repository must not yet be mistaken for a complete browsable source checkout. The source ZIP contains all2458 reviewed files and notices.

## Frozen artifact checksums

| Artifact | SHA256 |
| --- | --- |
| Linux amd64 runtime v20 | `7ceaa33e3093c25cdedc28dd5b7cc3e8fe5f245351361d5fba9dcec7a70ba0f3` |
| Complete source review ZIP v20 | `181e86d8257af9556a52e90a866001efceecc347c57d6b2c4722ea30de131df8` |

## License and provenance

GPL-3.0. See `LICENSE` and `XPANEL-NOTICE.md` in the complete source ZIP, `XPANEL-CHANGES.json` and `XPANEL-SHA256SUMS`. Upstream: [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui). Xpanel adaptation credit: **Eng.Abdelrahman Rabie 2026**. Upstream authorship and third-party licenses remain intact.
