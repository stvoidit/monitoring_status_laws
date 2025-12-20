import { ElMessageBox, ElNotification } from "element-plus";
import { getCurrentInstance } from "vue";
export const MessageAlert = (message: string) => {
    const appContext = getCurrentInstance()?.appContext;
    ElMessageBox({
        message: message,
        boxType: "alert",
        type: "error",
        showCancelButton: false,
        showClose: false,
        confirmButtonText: "Ok",
    }, appContext).catch(console.warn);
};
export const MessageSuccess = (message: string) => {
    const appContext = getCurrentInstance()?.appContext;
    ElMessageBox({
        message: message,
        boxType: "confirm",
        type: "success",
        showCancelButton: false,
        showClose: false,
        confirmButtonText: "Ok",
    }, appContext).catch(console.warn);
};

export const NotificationSaved = () => {
    const appContext = getCurrentInstance()?.appContext;
    ElNotification({
        title: "Сохранено",
        type: "success",
        duration: 2000,
        showClose: false,
    }, appContext);
};
