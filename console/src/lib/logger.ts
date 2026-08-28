const isProduction = import.meta.env.PROD;

const logger = {
    debug: (...args: any) => {
        if (!isProduction) {
            console.debug(...args);
        }
    },
    info: (...args: any) => {
        if (!isProduction) {
            console.info(...args);
        }
    },
    warn: (...args: any) => {
        console.warn(...args); // Warnings might be useful in production
    },
    error: (...args: any) => {
        console.error(...args); // Errors should always be logged
    },
};

export default logger;
