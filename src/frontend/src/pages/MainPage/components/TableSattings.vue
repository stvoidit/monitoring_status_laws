<template>
    <el-tooltip
        content="Настройка отображения полей таблицы"
        placement="bottom">
        <el-button
            id="menu-top-settings-table"
            @click="toggleVisible">
            <el-icon :size="18">
                <ElSetting />
            </el-icon>
        </el-button>
    </el-tooltip>
    <el-dialog
        v-model="dialogVisible"
        top="5vh"
        title="Отображение полей"
        width="25%">
        <VueDraggable
            v-model="model"
            :animation="150"
            :scroll="true"
            target=".sort-target">
            <TransitionGroup
                type="transition"
                class="sort-target"
                tag="span"
                name="fade">
                <div
                    v-for="field in model"
                    :key="field.prop"
                    class="drag-field">
                    <div class="icon-field">
                        <el-icon>
                            <ElRank />
                        </el-icon>
                    </div>
                    <div class="label-field">
                        {{ field.label }}
                    </div>
                    <div class="chbox-field">
                        <el-checkbox
                            :id="`setting-dialog-checkbox-${field.prop}`"
                            v-model.lazy="field.show"
                            size="large" />
                    </div>
                </div>
            </TransitionGroup>
        </VueDraggable>
    </el-dialog>
</template>

<script setup lang="ts">
import { Setting as ElSetting, Rank as ElRank } from "@element-plus/icons-vue";
import { VueDraggable } from "vue-draggable-plus";
import { ref, watchEffect, onBeforeMount } from "vue";
import type { ITableField } from "@/api/models";

const model = defineModel<ITableField[]>({ default: () => [] });
const dialogVisible = ref(false);
const toggleVisible = () => {
    dialogVisible.value = !dialogVisible.value;
};
onBeforeMount(() => {
    const loadFields = window.localStorage.getItem("fieldsSettings");
    if (Array.isArray(loadFields) && loadFields.length > 0) {
        model.value = JSON.parse(loadFields);
    }
});
watchEffect(() => {
    window.localStorage.setItem("fieldsSettings", JSON.stringify(model.value));
}, { flush: "post" });

</script>
<style scoped>
.el-dialog {
    text-align: left;
}
.zero-padding-cell {
    padding: 0 !important
}
.drag-field {
    border: 1px rgb(196, 188, 250) solid;
    display: flex;
    cursor: grab;
}
.icon-field {
    justify-content: center;
    width: 10%;
    align-items: center;
    display: flex;
}
.label-field {
    text-align: start;
    justify-content: start;
    align-items: center;
    display: flex;
    flex: auto;
}
.chbox-field {
    text-align: center;
    width: 15%;
    justify-content:
    center; display: flex;
}

.fade-move,
.fade-enter-active,
.fade-leave-active {
    transition: all 0.5s cubic-bezier(0.55, 0, 0.1, 1);
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
    transform: scaleY(0.01) translate(30px, 0);
}

.fade-leave-active {
    position: absolute;
}
.sort-target {
    padding: 0 1rem;
}
</style>
