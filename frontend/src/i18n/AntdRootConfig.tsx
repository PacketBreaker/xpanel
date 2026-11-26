import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { QuestionCircleOutlined } from '@ant-design/icons';
import { ConfigProvider } from 'antd';
import type { ConfigProviderProps } from 'antd/es/config-provider';
import type { Locale } from 'antd/es/locale';

// Xpanel ships English only: no extra antd locales are bundled, and undefined keeps
// antd's built-in en_US.
const antdLocaleLoaders: Record<string, () => Promise<{ default: Locale }>> = {};

let antdLocale: Locale | undefined;

export async function loadAntdLocale(language: string): Promise<void> {
  antdLocale = (await antdLocaleLoaders[language]?.())?.default;
}

// A Form.Item tooltip icon sits inside its <label>, so its "question-circle" name
// became part of every hinted field's name. FormField exposes the hint as a description.
const FORM_CONFIG: ConfigProviderProps['form'] = {
  tooltip: { icon: <QuestionCircleOutlined aria-hidden /> },
};

export function AntdRootConfig({ children }: { children: ReactNode }) {
  const { i18n } = useTranslation();
  const direction = /^(ar|fa)(-|$)/.test(i18n.language) ? 'rtl' : 'ltr';
  return (
    <ConfigProvider direction={direction} locale={antdLocale} form={FORM_CONFIG}>
      {children}
    </ConfigProvider>
  );
}
