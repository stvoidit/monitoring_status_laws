<template>
    <el-header style="padding: 0px">
        <el-row
            justify="space-between"
            style="padding: 0.5em; box-shadow: var(--el-box-shadow-light);">
            <el-col
                :span="3"
                style="padding-left:1.75em">
                <AddDocumentModal
                    v-if="appInfo.isAdmin||appInfo.isResponsible"
                    @add="fetchDocuments" />
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
                            @click="openTelegrmBot">
                            <el-icon :size="18">
                                <ElPromotion />
                            </el-icon>
                        </el-button>
                    </el-tooltip>
                    <el-tooltip
                        v-if="appInfo.isAdmin"
                        content="Настройки ролей пользователей"
                        placement="bottom">
                        <el-button
                            id="menu-top-settings-users"
                            @click="$router.push('/roles_settings')">
                            <el-icon :size="18">
                                <ElUser />
                            </el-icon>
                        </el-button>
                    </el-tooltip>
                    <el-tooltip
                        v-if="appInfo.isAdmin||appInfo.isResponsible"
                        content="Обновить все документы"
                        placement="bottom">
                        <el-button
                            id="menu-top-reload-documents"
                            @click="fetchUpdateAll">
                            <el-icon :size="18">
                                <ElMagicStick />
                            </el-icon>
                        </el-button>
                    </el-tooltip>
                    <el-tooltip
                        content="Скачать список документов в pdf"
                        placement="bottom">
                        <el-button
                            id="menu-top-download-pdf"
                            @click="downloadPDF">
                            <el-icon :size="18">
                                <el-icon><ElDocument /></el-icon>
                            </el-icon>
                        </el-button>
                    </el-tooltip>
                    <el-tooltip
                        content="Скачать список документов в xlsx"
                        placement="bottom">
                        <el-button
                            id="menu-top-download-documents"
                            @click="downloadDocuments">
                            <el-icon :size="18">
                                <ElDownload />
                            </el-icon>
                        </el-button>
                    </el-tooltip>
                    <TableSattings v-model="fields" />
                </el-button-group>
            </el-col>
        </el-row>
    </el-header>
</template>

<script setup lang="ts">
import useStore from "@/store";
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
import { appInfo } from "@/api";
import type { ITableField } from "@/api/models";
import { onBeforeMount } from "vue";

const {
    downloadDocuments,
    fetchDocuments,
    fetchUpdateAll,
    downloadPDF,
    loadFieldsSettings,
} = useStore();
onBeforeMount(loadFieldsSettings);
const fields = defineModel<ITableField[]>({ required: true });

/** функции-перемещения по страницам */
function openTelegrmBot() {
    const encodeUserID = window.btoa(appInfo.user?.id.toString() ?? "");
    const tgbotURL = new URL("https://t.me/");
    tgbotURL.pathname = appInfo.tg_bot_name;
    tgbotURL.searchParams.append("start", encodeUserID);
    window.open(tgbotURL, "_blank");
}

</script>
