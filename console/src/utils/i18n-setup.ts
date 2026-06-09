import i18next, { type ResourceLanguage } from 'i18next';
import { initReactI18next } from 'react-i18next';
import en from '../locales/en.json';

const getTranslations = (lng: ResourceLanguage) => ({
    translation: lng,
});

i18next
    .use(initReactI18next)
    .init({
        defaultNS: 'translation',
        lng: 'en',
        fallbackLng: 'en',
        debug: true,
        resources: {
            en: getTranslations(en),
        },
        react: {
            useSuspense: true,
        },
    })
    .catch((error) => {
        throw error;
    });

export { i18next };
