import 'i18next';
import EnTranslation from '../locales/en.json'; // For details, refer to https://www.i18next.com/overview/typescript#create-a-declaration-file

declare module 'i18next' {
    interface CustomTypeOptions {
        defaultNS: 'en';
        resources: {
            en: typeof EnTranslation;
        };
    }
}
