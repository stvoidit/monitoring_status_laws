<template>
    <el-col
        v-if="links"
        :span="7"
        class="col-padding">
        <el-text tag="b">
            {{ links.label }}:
        </el-text>
    </el-col>
    <el-col
        v-if="links"
        :span="17"
        class="col-padding">
        <ol class="ol-no-style">
            <el-text
                v-for="(file, index) in files"
                :key="file.href"
                tag="li">
                <el-space
                    direction="horizontal"
                    alignment="start">
                    <el-text>{{ index+1 }}.</el-text>
                    <el-link
                        :href="file.href"
                        type="primary"
                        target="_blank"
                        underline="never"
                        rel="noreferrer">
                        {{ file.name }}
                    </el-link>
                </el-space>
            </el-text>
        </ol>
    </el-col>
</template>

<script setup lang="ts">
import { computed } from "vue";
const {
    source,
    links = null,
} = defineProps<{
    source: string;
    links?: {
        label: string;
        files: {
            id: string;
            name: string;
            href: string;
        }[];
    };
}>();

const files = computed(() => {
    if (!links) return [];
    return links.files.map(
        ({ id, name }) => ({
            id, name, source,
            href: `/api/proxy_download?` + new URLSearchParams({ id, source, name }).toString(),
        }),
    );
});
</script>

<style scoped lang="css">
.ol-no-style {
    list-style-type: none;
    font-size: var(--el-font-size-base);
    margin: 0;
    padding: 0;
    padding-left: 1em;
}
.col-padding {
    padding-bottom:0.25em;
    padding-top:0.25em;
}
</style>
