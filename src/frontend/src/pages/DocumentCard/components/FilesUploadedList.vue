<template>
    <el-col :span="4">
        <el-text tag="b">
            Материалы:
        </el-text>
        <el-button
            v-if="userHasRights"
            style="margin-left: 1.5em;"
            size="small"
            @click="handleVisibility">
            <el-icon color="rgb(102.2, 177.4, 255)">
                <UploadFilled />
            </el-icon>
        </el-button>
    </el-col>
    <el-col
        :span="8"
        style="padding:0em">
        <el-upload
            id="upload-form"
            v-model:file-list="fileList"
            :data="{'documentID': documentId}"
            action="/api/files"
            :show-file-list="false"
            multiple
            drag
            with-credentials
            crossorigin="use-credentials"
            :on-remove="handleRemove"
            :on-success="onSuccess"
            :on-change="beforeUpload">
            <template #file="{ file }">
                <Teleport
                    defer
                    to="#files-scroll">
                    <li
                        class="el-upload-list__item is-success"
                        tabindex="0">
                        <div class="el-upload-list__item-info">
                            <a
                                class="el-upload-list__item-name"
                                :href="file.url">
                                <el-icon><Document /></el-icon>
                                <span>{{ file.name }}</span>
                            </a>
                        </div>
                        <label class="upload-icon-copystyle">
                            <el-icon
                                v-if="file.url"
                                id="check"
                                size="large"
                                color="rgb(133.4, 206.2, 97.4)">
                                <CircleCheck />
                            </el-icon>
                            <el-popconfirm
                                v-if="file.url"
                                title="Удалить файл?"
                                persistent
                                width="250"
                                confirm-button-text="удалить"
                                confirm-button-type="danger"
                                @confirm="handleRemove(file)">
                                <template #reference>
                                    <el-icon
                                        id="close"
                                        size="large"
                                        style="cursor: pointer;">
                                        <CircleClose />
                                    </el-icon>
                                </template>
                            </el-popconfirm>
                            <el-icon
                                v-else
                                size="large"
                                class="is-loading">
                                <Loading />
                            </el-icon>
                        </label>
                    </li>
                </Teleport>
            </template>
            <el-button type="primary">
                Загрузить
            </el-button>
        </el-upload>
        <el-scrollbar
            max-height="25vh"
            style="display: flex;flex-direction: column;">
            <ul
                id="files-scroll"
                class="el-upload-list el-upload-list--text"
                style="flex-direction: column-reverse;display: flex;">
                <li
                    v-for="file in fileList"
                    :key="file.uid"
                    class="el-upload-list__item is-success"
                    tabindex="0">
                    <div class="el-upload-list__item-info">
                        <a
                            class="el-upload-list__item-name"
                            :href="file.url">
                            <el-icon><Document /></el-icon>
                            <span>{{ file.name }}</span>
                        </a>
                    </div>
                    <label class="upload-icon-copystyle">
                        <el-icon
                            v-if="file.url"
                            id="check"
                            size="large"
                            color="rgb(133.4, 206.2, 97.4)">
                            <CircleCheck />
                        </el-icon>
                        <el-popconfirm
                            v-if="file.url"
                            title="Удалить файл?"
                            persistent
                            width="250"
                            confirm-button-text="удалить"
                            confirm-button-type="danger"
                            @confirm="handleRemove(file)">
                            <template #reference>
                                <el-icon
                                    id="close"
                                    size="large"
                                    style="cursor: pointer;">
                                    <CircleClose />
                                </el-icon>
                            </template>
                        </el-popconfirm>
                        <el-icon
                            v-else
                            size="large"
                            class="is-loading">
                            <Loading />
                        </el-icon>
                    </label>
                </li>
            </ul>
        </el-scrollbar>
    </el-col>
</template>

<script setup lang="ts">
import { GetFiles, DeleteFile } from "@/api";
import { ref, onBeforeMount } from "vue";
import { type UploadFile } from "element-plus";
import { UploadFilled, CircleCheck, CircleClose, Document, Loading } from "@element-plus/icons-vue";
interface AppUploadFile extends UploadFile {
    id?: string;
}
const { documentId } = defineProps<{
    documentId: string;
    userHasRights: boolean;
}>();
const fileList = ref<UploadFile[]>([]);
const edit = ref(true);
const fetchFiles = async () => {
    fileList.value = await GetFiles(documentId);
};
onBeforeMount(fetchFiles);

const handleRemove = async (uf: UploadFile) => {
    const fileID = (uf as AppUploadFile).id;
    if (fileID) {
        await DeleteFile(fileID).then(fetchFiles);
    }
};
const visibility = ref<"block" | "none">("none");
function handleVisibility() {
    if (visibility.value === "block") {
        visibility.value = "none";
        edit.value = false;
    } else {
        visibility.value = "block";
        edit.value = true;
    }
}
function onSuccess(response: { id: string; url: string }, uploadFile: AppUploadFile) {
    uploadFile.id = response.id;
    uploadFile.url = response.url;
}
function beforeUpload(uploadFile: UploadFile) {
    uploadFile.percentage = 0;
    uploadFile.status = "uploading";
}
</script>

<style>
#upload-form > .el-upload, .el-upload.is-drag {
    display: v-bind(visibility)
}
ol {
    font-size: 14px;
    font-weight: 500;
}
</style>
<style scoped>
.upload-icon-copystyle {
    align-items: center;
    height: 100%;
    justify-content: center;
    line-height: inherit;
    position: absolute;
    right: 5px;
    top: 0;
    transition: opacity var(--el-transition-duration);
    display: inline-flex;
}
.upload-icon-copystyle:not(:hover) > #close {
    display: none;
}
.upload-icon-copystyle:hover > #check {
    display: none;
}
</style>
