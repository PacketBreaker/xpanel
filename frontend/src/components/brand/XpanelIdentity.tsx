import { useTranslation } from 'react-i18next';
import './XpanelIdentity.css';

export function XpanelIdentity({ compact = false }: { compact?: boolean }) {
  const { t } = useTranslation();
  return <div className={`xpanel-identity${compact ? ' xpanel-identity-compact' : ''}`}>
    <span className="xpanel-monogram" aria-hidden="true">X</span>
    <span className="xpanel-wordmark" dir="ltr">Xpanel</span>
    {!compact && <span className="xpanel-edition">{t('xpanel.edition')}</span>}
  </div>;
}
export function XpanelCredit() {
  const { t } = useTranslation();
  return <footer className="xpanel-credit">
    <span>{t('xpanel.by')}</span> <bdi dir="ltr">Eng.Abdelrahman Rabie 2026</bdi>
    <span className="xpanel-provenance"><span>{t('xpanel.basedOn')}</span> <span className="xpanel-provenance-ltr" dir="ltr"><a href="https://github.com/MHSanaei/3x-ui" target="_blank" rel="noopener noreferrer"><bdi dir="ltr">3X-UI</bdi></a> <bdi dir="ltr">· GPLv3</bdi></span></span>
  </footer>;
}
