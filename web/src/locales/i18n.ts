import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

// 导入翻译文件
import enTranslations from './en.json';
import zhTranslations from './zh.json';
import jaTranslations from './ja.json';

const resources = {
  en: {
    translation: enTranslations,
  },
  zh: {
    translation: zhTranslations,
  },
  ja: {
    translation: jaTranslations,
  },
};

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'zh',
    lng: 'zh', // 默认语言设置为中文
    debug: false,

    interpolation: {
      escapeValue: false,
    },

    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      caches: ['localStorage'],
    },
  });

export default i18n;
