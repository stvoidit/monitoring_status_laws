<template>
    <el-button
        :text="text"
        :loading="loading"
        :disabled="disabled || loading"
        class="favorite-btn"
        @click="doFavorite">
        <el-icon
            v-show="!loading"
            class="el-rate is-active"
            :size="24">
            <ElStarFilled v-show="isFavorite" />
            <ElStar v-show="!isFavorite" />
        </el-icon>
    </el-button>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { DoFavorite } from "@/api";
import {
    Star as ElStar,
    StarFilled as ElStarFilled,
} from "@element-plus/icons-vue";
const {
    projectId,
    text,
    disabled,
} = defineProps<{
    projectId: string;
    text?: boolean;
    disabled?: boolean;
}>();
const loading = ref(false);
const isFavorite = defineModel<boolean>({ required: true });
const doFavorite = async () => {
    loading.value = true;
    const newValue = !isFavorite.value;
    try {
        await DoFavorite(projectId, newValue);
        isFavorite.value = newValue;
    } catch (reason) {
        alert(reason);
    } finally {
        loading.value = false;
    }
};
</script>
<style scoped src="element-plus/theme-chalk/el-rate.css"></style>
<style scoped>
.el-rate.is-active {
    color: var(--el-rate-fill-color) !important;
}
.text-center {
    text-align: center !important;
}
</style>
<style>
.el-table .cell:has(.favorite-btn) {
    text-align: center !important;
}
</style>
