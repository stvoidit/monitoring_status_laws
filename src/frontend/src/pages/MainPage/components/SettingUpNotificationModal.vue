<template>
    <el-tooltip
        content="Настройка уведомлений для телеграм бота"
        placement="bottom">
        <el-button
            id="menu-top-settings-table"
            @click="toggleVisible">
            <el-icon :size="18">
                <ElBell />
            </el-icon>
        </el-button>
    </el-tooltip>
    <el-dialog
        v-model="dialogVisible"
        center
        destroy-on-close
        title="Настройка уведомлений"
        width="35%">
        <el-radio-group
            v-model="appInfo!.ntype"
            size="large"
            class="vertical-radio"
            @change="changeNType">
            <el-radio
                v-for="opt in options"
                :key="opt.value"
                class="mb-2"
                border
                :aria-label="opt.label"
                :value="opt.value">
                {{ opt.label }}
            </el-radio>
        </el-radio-group>
    </el-dialog>
</template>

<script setup lang="ts">
import { Bell as ElBell } from "@element-plus/icons-vue";
import { ref } from "vue";
import {
    appInfo,
    ChangeNotificationType,
} from "@/api";
import { isAxiosError } from "axios";
import { ElMessageBox, ElNotification } from "element-plus";
/** Изменение типа уведомлений */
const changeNType = async (ntype: string | number | boolean | undefined) => {
    try {
        const response = await ChangeNotificationType(ntype as number);
        if (response.status === 201) {
            ElNotification({
                title: "Сохранено",
                type: "success",
                showClose: false,
                duration: 750,
            });
        }
    } catch (error) {
        console.error(error);
        if (isAxiosError(error)) {
            await ElMessageBox.alert(error.response?.data.error as string, "Ошибка").catch(console.warn);
        }
    };
};

const dialogVisible = ref(false);
const toggleVisible = () => {
    dialogVisible.value = !dialogVisible.value;
};
const options = [
    {
        label: "Получать уведомления по всем документам",
        value: 0,
    },
    {
        label: "Получать уведомления только по избранным документам",
        value: 1,
    },
    {
        label: "Получать уведомления только по важным документам (1 и 2 приоритет)",
        value: 2,
    },
];
</script>
<style scoped>
.vertical-radio {
    flex-direction: column;
    align-items: flex-start;
    flex-wrap: nowrap;
    justify-content: flex-start;
    width: 100%;
}
.el-radio.el-radio--large {
    width: 100%;
}
</style>
