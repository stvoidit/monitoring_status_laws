<template>
    <el-row
        align="middle"
        class="shadow-box "
        style="padding:1em">
        <el-col :span="20">
            <h6>Журнал стадий рассмотрения</h6>
        </el-col>
        <el-col
            :span="4"
            style="text-align:end">
            <el-tooltip
                content="Сортировка по дате изменения"
                placement="bottom">
                <el-button
                    size="small"
                    plain
                    :icon="directionIcon"
                    @click="handleChangeDirection" />
            </el-tooltip>
        </el-col>
        <el-col :span="24">
            <el-table
                border
                stripe
                size="small"
                class-name="normalize-table"
                :data="computedStages"
                style="width: 100%">
                <el-table-column
                    width="130"
                    label="Дата"
                    prop="Date" />
                <el-table-column
                    label="Событие"
                    prop="header" />
                <el-table-column
                    label="Статус"
                    prop="decision" />
            </el-table>
        </el-col>
    </el-row>
</template>

<script setup lang="ts">
import { computed } from "vue";
import useSortDirection from "@/utils/useSortDirection";
import { JournalRow } from "@/api/models";

const { stages = [] } = defineProps<{
    stages?: JournalRow[];
}>();
const {
    sortDict,
    directionIcon,
    handleChangeDirection,
} = useSortDirection();

const computedStages = computed(() => {
    if (!stages.length) return [];
    return [ ...stages ].sort((a, b) => a.date < b.date ? -1 * sortDict.value : 1 * sortDict.value);
});

</script>
<style scoped>
.shadow-box {
    box-shadow: var(--el-box-shadow-lighter);
    margin-top: 1em;
}
</style>
