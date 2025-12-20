<template>
    <el-config-provider :locale="locale">
        <el-container
            v-if="inited"
            direction="vertical">
            <router-view />
        </el-container>
    </el-config-provider>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { init } from "@/api";
import locale from "@/utils/configLocale";

const inited = ref(false);
const initFn = () => {
    init().then(() => {
        inited.value = true;
    }).catch(console.error);
};
onMounted(initFn);

</script>
<style src="@/assets/fix-table.css"></style>
<style src="@/assets/moscow-fonts.css"></style>
<style lang="css">

:root {
    --el-text-color: var(--el-color-black);
}
.el-pagination .el-select {
    width: 4rem !important;
}
.el-text {
    --el-text-color: var(--el-color-black);
}
#app {
    background-color: #ffffff;
}
</style>
