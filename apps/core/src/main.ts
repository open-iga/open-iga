import { getAppConfig } from '@/common/config-manager.ts';
import { createRuntimeRemotes } from '@/remote/create-runtime-remotes.ts';
import { createRuntimeApplication } from '@/application/create-runtime-application.ts';
import { createRuntimeApi } from '@/api/create-runtime-api.ts';
import { logger } from '@/common/logger.ts';

const banner =
    '\n' +
    ' ██████╗ ██████╗ ███████╗███╗   ██╗    ██╗ ██████╗  █████╗ \n' +
    '██╔═══██╗██╔══██╗██╔════╝████╗  ██║    ██║██╔════╝ ██╔══██╗\n' +
    '██║   ██║██████╔╝█████╗  ██╔██╗ ██║    ██║██║  ███╗███████║\n' +
    '██║   ██║██╔═══╝ ██╔══╝  ██║╚██╗██║    ██║██║   ██║██╔══██║\n' +
    '╚██████╔╝██║     ███████╗██║ ╚████║    ██║╚██████╔╝██║  ██║\n' +
    ' ╚═════╝ ╚═╝     ╚══════╝╚═╝  ╚═══╝    ╚═╝ ╚═════╝ ╚═╝  ╚═╝\n' +
    '                                                           ';

const main = async () => {
    const appConfig = getAppConfig();

    const runtimeRemotes = createRuntimeRemotes({ appConfig });
    const runtimeApplication = createRuntimeApplication({ runtimeRemotes });
    const runtimeApi = createRuntimeApi({ runtimeApplication, appConfig });

    runtimeApi.listen(appConfig.port, () => {
        logger.info(banner);
        logger.info(`API is running on port ${appConfig.port} in ${appConfig.env} environment 🚀`);
    });
};

await main();
