<template>
    <el-header style="padding: 0px">
        <el-row
            justify="space-between"
            class="menu-row">
            <el-col
                :span="3"
                style="padding-left:1.75em">
                <AddDocumentModal
                    v-if="appInfo.isAdmin||appInfo.isResponsible"
                    @add="emitAddDocument" />
            </el-col>
            <el-col
                :span="21">
                <el-button-group style="float:right">
                    <SettingUpNotificationModal />
                    <el-tooltip
                        content="Телеграм бот"
                        placement="bottom">
                        <el-button
                            id="menu-top-settings-users"
                            :icon="ElPromotion"
                            @click="openTelegrmBot" />
                    </el-tooltip>
                    <el-tooltip
                        v-if="appInfo.isAdmin"
                        content="Настройки ролей пользователей"
                        placement="bottom">
                        <el-button
                            id="menu-top-settings-users"
                            :icon="ElUser"
                            @click="$router.push('/roles_settings')" />
                    </el-tooltip>
                    <el-tooltip
                        v-if="appInfo.isAdmin||appInfo.isResponsible"
                        content="Обновить все документы"
                        placement="bottom">
                        <el-button
                            id="menu-top-reload-documents"
                            :icon="ElMagicStick"
                            @click="emitUpdateAll" />
                    </el-tooltip>
                    <el-tooltip
                        content="Скачать список документов в pdf"
                        placement="bottom">
                        <el-button
                            id="menu-top-download-pdf"
                            :icon="ElDocument"
                            @click="emitDownloadPDF" />
                    </el-tooltip>
                    <el-tooltip
                        content="Скачать список документов в xlsx"
                        placement="bottom">
                        <el-button
                            id="menu-top-download-documents"
                            :icon="ElDownload"
                            @click="emitDownloadXLSX" />
                    </el-tooltip>
                    <TableSattings v-model="fields" />
                </el-button-group>
            </el-col>
        </el-row>
    </el-header>
</template>

<script setup lang="ts">
import {
    Promotion as ElPromotion,
    User as ElUser,
    MagicStick as ElMagicStick,
    Download as ElDownload,
    Document as ElDocument,
} from "@element-plus/icons-vue";
import AddDocumentModal from "./AddDocumentModal.vue";
import TableSattings from "./TableSattings.vue";
import SettingUpNotificationModal from "./SettingUpNotificationModal.vue";
import { appInfo, openTelegrmBot } from "@/api";
import type { ITableField } from "@/api/models";

const emit = defineEmits<{
    downloadXLSX: [];
    downloadPDF: [];
    addDocument: [];
    updateAll: [];
}>();
function emitDownloadPDF() {
    emit("downloadPDF");
}
function emitDownloadXLSX() {
    emit("downloadXLSX");
}
const emitAddDocument = () => {
    emit("addDocument");
};
const emitUpdateAll = () => {
    emit("updateAll");
};

const fields = defineModel<ITableField[]>({ required: true });

</script>
<style scoped>
.menu-row {
    padding: 0.5em;
    box-shadow: var(--el-box-shadow-light);
}
</style>
