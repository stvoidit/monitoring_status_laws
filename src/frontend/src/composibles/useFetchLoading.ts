import { ref, toRef, type Ref } from "vue";

const useAbortController = () => {
    let abortController: AbortController | null = null;
    const makeSignal = () => {
        abortController?.abort();
        abortController = new AbortController();
        return {
            signal: abortController.signal,
            abort: () => {
                abortController?.abort();
                abortController = null;
            },
        };
    };
    return makeSignal;
};

const useFetchLoading = <T>(fetchFn: (signal?: AbortSignal) => Promise<T>, value: Ref<T>) => {
    const loading = ref(true);
    const dataValue = toRef(value);
    const makeSignal = useAbortController();
    const doFetch = async () => {
        loading.value = true;
        const { signal, abort } = makeSignal();
        try {
            dataValue.value = await fetchFn(signal).finally(abort);
        } catch (error) {
            console.error(error);
            throw error;
        } finally {
            loading.value = false;
        }
    };
    return {
        loading,
        value,
        doFetch,
    };
};
export default useFetchLoading;

const parseResponseError = async <E = { message: string }>(r: Response) => {
    if (r.headers.get("content-type")?.includes("json")) {
        return await r.json() as E;
    }
    return await r.text();
};

export type methods = "get" | "post" | "put" | "patch" | "delete";
export const createFetchLoading = <T>(method: methods, url: string, value: Ref<T>) => {
    const dataValue = toRef(value);
    const fetchFn = async (signal?: AbortSignal) => {
        const response = await fetch(url, {
            keepalive: true,
            method,
            signal,
        });
        if (!response.ok) {
            const err = await parseResponseError(response);
            if (typeof err === "object" && "message" in err) {
                throw new Error(err.message);
            }
            throw new Error(err);
        }
        const body = await response.json();
        return body as T;
    };
    return useFetchLoading<T>(fetchFn, dataValue);
};
