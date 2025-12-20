<template>
    <el-container
        v-if="doc"
        direction="vertical">
        <HeaderCard
            :id="doc.id"
            v-model:is-favorite="doc.is_favorite"
            v-model:editing="editing"
            v-model="doc"
            :project="doc.project"
            :source-url="doc.DocumentSourceURL"
            :percent="doc.percent"
            :review-stage="doc.ReviewStage"
            :update-date="doc.Updated"
            :law-type="doc.LawType"
            :is-draft="doc.is_draft"
            :archive="doc.archive"
            :user-has-rights="userHasRights"
            :loading="loading"
            @changed="fetchDocument"
            @reload="fetchUpdate"
            @delete="deleteDocument"
            @source="setSource" />
        <el-main style="padding-top: 1em;">
            <el-row class="shadow-box">
                <el-col
                    v-if="!doc.is_draft"
                    :span="12">
                    <h3 class="subtitle">
                        Общая информация
                    </h3>
                    <el-divider />
                    <el-row>
                        <template
                            v-for="field in documentFields.baseInformation"
                            :key="field.prop">
                            <el-col
                                :span="field.span1"
                                class="col-padding">
                                <el-text tag="b">
                                    {{ field.name }}:
                                </el-text>
                            </el-col>
                            <el-col
                                :span="field.span2"
                                class="col-padding">
                                <el-text>{{ doc[field.prop] }}</el-text>
                            </el-col>
                        </template>
                    </el-row>
                </el-col>
                <el-col
                    v-if="!doc.is_draft"
                    :span="12">
                    <h3 class="subtitle">
                        Рассмотрение законопроекта
                    </h3>
                    <el-divider />
                    <el-row>
                        <template
                            v-for="field in documentFields.lawReview"
                            :key="field.prop">
                            <el-col
                                :span="field.span1"
                                class="col-padding">
                                <el-text tag="b">
                                    {{ field.name }}:
                                </el-text>
                            </el-col>
                            <el-col
                                :span="field.span2"
                                class="col-padding">
                                <el-input
                                    v-if="editing && field.editable"
                                    :id="`edit-card-field-textarea-${field.prop}`"
                                    v-model="doc[field.prop]"
                                    type="textarea"
                                    @change="onChangeUpdate">
                                    <template #prepend />
                                </el-input>
                                <el-text v-else>
                                    {{ doc[field.prop] }}
                                </el-text>
                            </el-col>
                        </template>
                        <ExternalFiles
                            :links="doc.Files2read"
                            :source="doc.source" />
                        <ExternalFiles
                            :links="doc.Files3read"
                            :source="doc.source" />
                        <ExternalFiles
                            v-for="partFiles in doc.FilesAll"
                            :key="partFiles.label"
                            :source="doc.source"
                            :links="partFiles" />
                    </el-row>
                </el-col>
                <el-col style="padding-top:1em">
                    <h3 class="subtitle">
                        Дополнительная информация
                    </h3>
                    <el-divider />
                    <el-row>
                        <!-- костыль, т.к. захотели иметь редактируемое поле не в разделе additionInformation -->
                        <template v-if="!doc.is_draft">
                            <el-col
                                :span="4"
                                class="col-padding">
                                <el-text tag="b">
                                    Актуализированный статус:
                                </el-text>
                            </el-col>
                            <el-col
                                :span="20"
                                class="col-padding">
                                <el-input
                                    v-if="editing"
                                    id="edit-card-field-textarea-actual_status"
                                    v-model="doc.actual_status"
                                    type="textarea"
                                    autosize
                                    @change="onChangeUpdate" />
                                <AutoLinkerText
                                    v-else
                                    id="show-card-field-textarea-actual_status"
                                    :text="doc.actual_status" />
                            </el-col>
                        </template>
                        <template
                            v-for="field in documentFields.additionInformation"
                            :key="field.prop">
                            <el-col
                                :span="field.span1"
                                class="col-padding">
                                <el-text tag="b">
                                    {{ field.name }}:
                                </el-text>
                            </el-col>
                            <el-col
                                :span="field.span2"
                                class="col-padding">
                                <template v-if="editing">
                                    <el-select
                                        v-if="field.isSelect"
                                        v-model="doc[field.prop]"
                                        @change="onChangeUpdate">
                                        <el-option
                                            v-for="priority in [1,2,3]"
                                            :key="priority"
                                            :label="priority"
                                            :value="priority" />
                                    </el-select>
                                    <el-date-picker
                                        v-else-if="field.isDatepicker"
                                        v-model="doc[field.prop]"
                                        type="date"
                                        format="DD.MM.YYYY"
                                        @change="onChangeUpdate" />
                                    <el-input
                                        v-else
                                        :id="`edit-card-field-textarea-${field.prop}`"
                                        v-model="doc[field.prop]"
                                        type="textarea"
                                        autosize
                                        @change="onChangeUpdate" />
                                </template>
                                <AutoLinkerText
                                    v-else
                                    :id="`show-card-field-textarea-${field.prop}`"
                                    :text="doc[field.prop]" />
                            </el-col>
                        </template>
                    </el-row>
                </el-col>
                <FilesUploadedList
                    :document-id="documentId"
                    :user-has-rights="userHasRights" />
            </el-row>
            <StagesJournal
                v-if="!doc.is_draft"
                :stages="doc.journal" />
            <LogsJournal :logs="logs" />
        </el-main>
    </el-container>
</template>

<script setup lang="ts">
import AutoLinkerText from "./components/AutoLinkerText";
import { onBeforeMount } from "vue";
import { ElText } from "element-plus";
import HeaderCard from "./components/HeaderCard.vue";
import LogsJournal from "./components/LogsJournal.vue";
import StagesJournal from "./components/StagesJournal.vue";
import FilesUploadedList from "./components/FilesUploadedList.vue";
import ExternalFiles from "./components/ExternalFiles.vue";
import useState from "./state";

const { documentId } = defineProps<{
    documentId: string;
}>();
const {
    documentFields,
    doc, logs,
    editing,
    loading,
    userHasRights,
    fetchDocument,
    fetchUpdate,
    onChangeUpdate,
    deleteDocument,
    setSource,
} = useState(documentId);

onBeforeMount(fetchDocument);

</script>
<style scoped lang="css">
.col-padding {
    padding-bottom:0.25em;
    padding-top:0.25em;
}
</style>
<style lang="css">

.confirm-dialog .el-dialog__body {
    padding: 0.25em 0 !important;
}
.buttton-danger {
    background-color: var(--el-color-error);
}
.button-warning {
    background-color: var(--el-color-warning);
}
.shadow-box {
    box-shadow: var(--el-box-shadow-lighter);
    margin-top: 1em;
    padding: 0 1em 1em 1em;
}
.subtitle {
    margin: 0.5em 0.25em 0.25em 0 !important;
}

</style>
