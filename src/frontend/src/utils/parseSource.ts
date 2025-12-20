export const sources = [
    "regulation.gov.ru",
    "sozd.duma.gov.ru",
] as const;

const regulationRgx = RegExp(/projects\/(\d+)/);

export const parseSourceURL = (link: string) => {
    const payload = {
        id: "",
        source: "",
    };
    try {
        const url = new URL(link);
        const sourceHost = sources.find(s => s === url.hostname);
        payload.source = sourceHost ?? "";
        switch (payload.source) {
            case "regulation.gov.ru":
                const match = regulationRgx.exec(url.pathname);
                payload.id = match?.at(1) ?? "";
                break;
            case "sozd.duma.gov.ru":
                payload.id = url.pathname.replace("/bill/", "");
                break;
            default:
                payload.id = "";
                break;
        }
    } catch (error) {
        console.warn(error);
        payload.id = "";
        payload.source = "";
    }
    return payload;
};
