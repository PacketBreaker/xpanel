import { render, screen, cleanup } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { readyI18n } from '@/i18n/react';
import { XpanelIdentity, XpanelCredit } from '@/components/brand/XpanelIdentity';
afterEach(cleanup);
describe('Xpanel identity and provenance', () => {
  it.each(['en-US', 'ar-EG'])('preserves credit and upstream attribution in %s', async (language) => {
    document.cookie = `lang=${language}; path=/`;
    await readyI18n();
    render(<><XpanelIdentity /><XpanelCredit /></>);
    expect(screen.getByText('Xpanel')).toBeTruthy();
    expect(screen.getByText('Eng.Abdelrahman Rabie 2026')).toBeTruthy();
    expect(screen.getByRole('link', { name: '3X-UI' }).getAttribute('href')).toBe('https://github.com/MHSanaei/3x-ui');
    expect(document.documentElement.dir).toBe(language === 'ar-EG' ? 'rtl' : 'ltr');
  });
});
