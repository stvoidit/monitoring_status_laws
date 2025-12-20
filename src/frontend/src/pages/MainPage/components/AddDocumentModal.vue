<template>
    <el-button
        id="menu-top-add-project"
        type="success"
        :plain="!dialogVisible"
        @click="toggleVisible(true)">
        Добавить проект
    </el-button>
    <el-dialog
        v-model="dialogVisible"
        title="Создание карточки законопроекта"
        width="60%"
        teleported
        append-to-body
        @close="handlReset">
        <el-form
            ref="refForm"
            label-position="top"
            :model="newDocument"
            label-width="280px"
            :rules="formRules">
            <el-row :gutter="20">
                <el-col :span="12">
                    <el-form-item
                        v-if="requiredFields"
                        label="Источник:"
                        prop="source">
                        <el-select
                            id="card-istochnik"
                            v-model="newDocument.source"
                            fit-input-width
                            placeholder="Выберите источник">
                            <el-option
                                v-for="source in sources"
                                :key="source"
                                :label="source"
                                :value="source" />
                        </el-select>
                    </el-form-item>
                </el-col>
                <el-col :span="12">
                    <el-form-item
                        label="Идентификатор:"
                        prop="id">
                        <el-input
                            id="card-identificator"
                            v-model="newDocument.id" />
                    </el-form-item>
                </el-col>
            </el-row>

            <el-row
                v-if="requiredFields"
                :gutter="20">
                <el-col :span="12">
                    <el-form-item
                        label="Ссылка на законопроект:"
                        prop="sourceURL">
                        <el-input
                            id="card-law-url"
                            v-model="newDocument.sourceURL"
                            @input="handleParseSource" />
                    </el-form-item>
                </el-col>
                <el-col :span="6">
                    <el-form-item label="Приоритет:">
                        <el-select v-model="newDocument.priority">
                            <el-option
                                v-for="priority in [1,2,3]"
                                :key="priority"
                                :label="priority"
                                :value="priority" />
                        </el-select>
                    </el-form-item>
                </el-col>
                <el-col :span="6">
                    <el-form-item label="">
                        <el-checkbox
                            v-model="newDocument.is_draft"
                            label="Черновик:" />
                    </el-form-item>
                </el-col>
            </el-row>
            <el-form-item
                label="Краткое название:"
                prop="short_label">
                <el-input
                    id="card-short-text"
                    v-model="newDocument.short_label"
                    type="textarea"
                    :autosize="{minRows:1}" />
            </el-form-item>
            <el-form-item label="Вид налога (сбора):">
                <el-input
                    id="card-type-tax"
                    v-model="newDocument.tax_type" />
            </el-form-item>
            <el-form-item label="Область регулирования:">
                <el-input
                    id="card-oblast"
                    v-model="newDocument.scope" />
            </el-form-item>
            <el-form-item label="Краткое содержание:">
                <el-input
                    id="card-short-text"
                    v-model="newDocument.desc"
                    type="textarea"
                    :autosize="{minRows:2}" />
            </el-form-item>

            <el-form-item label="Примечания:">
                <el-input
                    id="card-note"
                    v-model="newDocument.note"
                    type="textarea"
                    :autosize="{minRows:2}" />
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button
                id="menu-top-close-dialog"
                :disabled="loading"
                @click="toggleVisible(false)">
                Закрыть
            </el-button>
            <el-button
                v-if="loading"
                type="warning"
                @click="handleCancelRequest">
                Остановить
            </el-button>
            <el-button
                id="menu-top-add-document"
                :loading="loading"
                type="primary"
                :disabled="formNotReady"
                @click="addNewDocument">
                Добавить
            </el-button>
        </template>
    </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, shallowRef, useTemplateRef } from "vue";
import { parseSourceURL, sources } from "@/utils/parseSource";
import { AddNewDocument, isCanceledError } from "@/api";
import { ElNotification } from "element-plus";
import { AxiosError } from "axios";
const emit = defineEmits<{
    add: [];
}>();
const loading = ref(false);
const dialogVisible = ref(false);
const toggleVisible = (value?: boolean) => {
    dialogVisible.value = value ?? !dialogVisible.value;
};
const newDocument = reactive({
    source: "",
    id: "",
    short_label: "",
    desc: "",
    scope: "",
    tax_type: "",
    priority: 3,
    note: "",
    sourceURL: "",
    is_draft: false,
});

const refForm = useTemplateRef("refForm");
const handlReset = () => {
    if (loading.value) {
        return;
    }
    newDocument.source = "";
    newDocument.id = "";
    newDocument.desc = "";
    newDocument.short_label = "";
    newDocument.tax_type = "";
    newDocument.priority = 3;
    newDocument.note = "";
    newDocument.scope = "";
    newDocument.sourceURL = "";
    newDocument.is_draft = false;
    refForm.value?.resetFields();
};

const abortController = shallowRef<AbortController>();
const handleCancelRequest = () => {
    if (abortController.value) {
        abortController.value.abort();
    }
};
const newController = () => {
    handleCancelRequest();
    abortController.value = new AbortController();
    return abortController.value.signal;
};

const addNewDocument = async () => {
    loading.value = true;
    try {
        const signal = newController();
        await AddNewDocument(newDocument, { signal });
        emit("add");
        ElNotification({
            title: "Добавлено",
            type: "success",
            showClose: false,
            duration: 750,
        });
    } catch (reason) {
        console.log(reason);
        if (isCanceledError(reason)) {
            return;
        } else {
            alert((reason as AxiosError<{ error: string }>).response?.data.error);
        }
    } finally {
        loading.value = false;
        handlReset();
        toggleVisible(false);
    };
};
const formNotReady = computed(() => {
    if (newDocument.is_draft) {
        return newDocument.short_label.trim() === "";
    } else {
        return (newDocument.id === "" || newDocument.source === "");
    }
});

const handleParseSource = (link: string) => {
    const { id, source } = parseSourceURL(link);
    newDocument.id = id;
    newDocument.source = source;
};

const requiredFields = computed(() => !newDocument.is_draft);
const formRules = reactive({
    source: [ { required: requiredFields } ],
    id: [ { required: requiredFields } ],
    sourceURL: [ { required: requiredFields } ],
    short_label: [ { required: newDocument.is_draft } ],
});
</script>

<style>
.el-dialog__body {
    word-break: normal !important;
}
.el-form-item__error {
    position: initial !important;
}
</style>
