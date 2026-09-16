const isProduction = import.meta.env.PROD;

const logger = {
    debug: (...args: unknown[]) => {
        if (!isProduction) {
            console.debug(...args);
        }
    },
    info: (...args: unknown[]) => {
        if (!isProduction) {
            console.info(...args);
        }
    },
    warn: (...args: unknown[]) => {
        console.warn(...args); // Warnings might be useful in production
    },
    error: (...args: unknown[]) => {
        console.error(...args); // Errors should always be logged
    },
};

export default logger;
