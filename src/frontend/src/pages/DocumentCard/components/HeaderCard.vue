<template>
    <el-header style="padding: 0px">
        <el-row
            justify="center"
            align="middle"
            class="shadow-box header-row">
            <el-col
                :span="8"
                style="padding-left:0.57em">
                <el-button-group class="ml-4">
                    <el-tooltip
                        content="Вернуться в реестр">
                        <el-button
                            id="edit-card-return"
                            :icon="ElBack"
                            @click="$router.push('/')" />
                    </el-tooltip>
                    <el-tooltip
                        v-if="!isDraft"
                        content="Последняя проверка"
                        placement="bottom">
                        <el-button
                            style="cursor: default;"
                            text>
                            {{ updateDate }}
                        </el-button>
                    </el-tooltip>
                    <el-tooltip
                        v-if="userHasRights"
                        content="Редактировать документ"
                        placement="bottom">
                        <el-button
                            id="edit-card-edit-document"
                            :disabled="loading"
                            :type="editingType"
                            :icon="ElEditPen"
                            @click="editing = !editing" />
                    </el-tooltip>
                    <el-tooltip
                        v-if="userHasRights"
                        content="Удалить документ"
                        placement="bottom">
                        <el-button
                            id="edit-card-delete-document"
                            :disabled="editing"
                            :icon="ElClose"
                            @click="emit('delete', id)" />
                    </el-tooltip>
                    <el-tooltip
                        v-if="userHasRights"
                        :content="draftTooltip"
                        placement="bottom">
                        <el-button
                            v-if="!isDraft"
                            id="edit-card-reload-document"
                            :loading="loading"
                            :disabled="loading || editing"
                            @click="emit('reload', id)">
                            <el-icon v-show="!loading">
                                <ElMagicStick />
                            </el-icon>
                        </el-button>
                        <el-button
                            v-else
                            id="edit-card-reload-document"
                            :icon="ElMagicStick"
                            @click="handleSetSource" />
                    </el-tooltip>
                    <el-tooltip
                        :content="archiveTooltip"
                        placement="bottom">
                        <el-button
                            :disabled="editing || loading"
                            :icon="ElTakeawayBox"
                            @click="handleToArchive" />
                    </el-tooltip>
                    <FavoriteButton
                        v-model="isFavorite"
                        :disabled="editing"
                        :project-id="id" />
                    <el-button
                        :icon="ElShare"
                        @click="shareLink" />
                </el-button-group>
            </el-col>
            <el-col
                style="text-align:center;"
                :span="8">
                <el-text v-if="!isDraft">
                    {{ lawType }}: <el-link
                        id="edit-card-document-link"
                        type="primary"
                        target="_blank"
                        :href="sourceUrl">
                        {{ project }}
                    </el-link>
                </el-text>
            </el-col>
            <el-col :span="8">
                <el-row>
                    <el-col
                        v-if="!isDraft && !archive"
                        :span="20"
                        style="text-align:end;cursor: default;">
                        <el-tag type="primary">
                            {{ reviewStage }}
                        </el-tag>
                    </el-col>
                    <el-col
                        v-if="!isDraft && !archive"
                        :offset="0"
                        :span="4"
                        style="text-align:center;cursor: default;">
                        <el-tag type="primary">
                            {{ percent }}
                        </el-tag>
                    </el-col>
                    <el-col
                        v-if="archive"
                        style="text-align:end;cursor: default;">
                        <el-tag
                            hit
                            class="tag-archive"
                            color="#a493933d">
                            Архив
                        </el-tag>
                    </el-col>
                </el-row>
            </el-col>
        </el-row>
    </el-header>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { parseSourceURL } from "@/utils/parseSource";
import { ElNotification, ElMessageBox, ElText } from "element-plus";
import FavoriteButton from "@/components/FavoriteButton.vue";
import {
    Back as ElBack,
    EditPen as ElEditPen,
    Close as ElClose,
    MagicStick as ElMagicStick,
    Share as ElShare,
    TakeawayBox as ElTakeawayBox,
} from "@element-plus/icons-vue";
import {
    ChangeArchiveStatus,
} from "@/api";

const {
    id,
    archive,
    isDraft,
} = defineProps<{
    id: string;
    project: string;
    isDraft: boolean;
    lawType: "Закон" | "Законопроект";
    reviewStage: string;
    userHasRights: boolean;
    loading: boolean;
    archive: boolean;
    updateDate: string;
    percent: string;
    sourceUrl: string;
}>();

const emit = defineEmits<{
    changed: [id: string];
    reload: [id: string];
    delete: [id: string];
    source: [
        source: {
            id: string;
            source: string;
        },
    ];
}>();

const editing = defineModel<boolean>("editing", { required: true });
const isFavorite = defineModel<boolean>("isFavorite", { required: true });

const variants = (isArchive: boolean) => {
    if (isArchive) {
        return {
            text: "Вернуть проект из архива?",
            boxType: "info",
            buttonClass: "",
            notificationText: "Проект возвращен из архива",
        } as const;
    } else {
        return {
            text: "Отправить проект в архив?",
            boxType: "warning",
            buttonClass: "button-warning",
            notificationText: "Проект перемещен в архива",
        } as const;
    }
};
const editingType = computed(() => {
    return editing.value ? "primary" : "default";
});

const handleToArchive = async () => {
    const variant = variants(archive);
    try {
        const result = await ElMessageBox.confirm(variant.text, {
            confirmButtonText: "Да",
            cancelButtonText: "Отмена",
            type: variant.boxType,
            confirmButtonClass: variant.buttonClass,
        });
        if (result === "confirm") {
            await ChangeArchiveStatus(id);
            ElNotification({
                title: "",
                message: variant.notificationText,
                type: variant.boxType,
                duration: 1000,
                showClose: false,
            });
            emit("changed", id);
        }
    } catch (error) {
        if (error instanceof Error) {
            console.error(error);
        }
    }
};

const shareLink = () => {
    const el = document.createElement("input");
    el.value = window.location.href;
    el.readOnly = true;
    document.body.appendChild(el);
    el.select();
    // eslint-disable-next-line @typescript-eslint/no-deprecated
    document.execCommand("copy");
    document.body.removeChild(el);
    ElNotification({
        title: "",
        message: "ссылка скопирована",
        type: "info",
        duration: 1000,
        showClose: false,
    });
};

const handleSetSource = async () => {
    const message = "Введите ссылку на документ";
    const title = "Добавление источника документа";
    try {
        const result = await ElMessageBox.prompt(message, title, {
            confirmButtonText: "Добавить",
            cancelButtonText: "Закрыть",
            // inputPattern: RegExp(/https:\/\/[]/),
            // inputValidator: (valueText) => {
            //     try {
            //         parseSourceURL(valueText);
            //     } catch (error) {
            //         return (error as Error).message;
            //     }
            //     return true;
            // },
        });
        if (result.action !== "confirm") return;
        const payload = parseSourceURL(result.value);
        emit("source", payload);
    } catch (error) {
        if (error instanceof Error) {
            console.warn(error);
        }
    }
};

const archiveTooltip = computed(() => {
    return archive ? "Вернуть проект из архива" : "Отправить проект в архив";
});
const draftTooltip = computed(() => {
    return isDraft ? "Добавить ссылку на источник" : "Обновить документ из источника";
});

</script>

<style scoped lang="css">
.header-row {
    padding-top: 0.75em !important;
    padding-bottom: 0.75em !important;
    margin-left: 1em !important;
    margin-right: 1em !important;
    margin-top: 0;
    margin-bottom: 0;
}
.shadow-box {
    box-shadow: var(--el-box-shadow-lighter);
    margin-top: 1em;
    padding: 0 1em 1em 1em;
}
.tag-archive {
    margin-right: 1em;
    cursor: default;
    color: rgb(115.2,117.6,122.4);
    border-color: rgb(199.5,201,204);
}
</style>
