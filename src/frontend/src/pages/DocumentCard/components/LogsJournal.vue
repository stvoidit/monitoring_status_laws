<template>
    <el-row
        align="middle"
        class="shadow-box"
        style="padding:1em">
        <el-col :span="20">
            <h6>Журнал логирования</h6>
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
                :data="computedLogs"
                style="width: 100%">
                <el-table-column
                    width="150"
                    label="Дата"
                    prop="Created" />
                <el-table-column
                    width="250"
                    label="Пользователь"
                    prop="user.shortname" />
                <el-table-column
                    width="250"
                    label="Поле"
                    prop="ReadebleFieldName" />
                <el-table-column
                    label="До"
                    prop="Before" />
                <el-table-column
                    label="После"
                    prop="After" />
            </el-table>
        </el-col>
    </el-row>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { type LogRow } from "@/api/models";
import useSortDirection from "@/utils/useSortDirection";

const { logs = [] } = defineProps<{
    logs?: LogRow[];
}>();

const {
    sortDict,
    directionIcon,
    handleChangeDirection,
} = useSortDirection();

const computedLogs = computed(() => {
    if (!logs.length) return [];
    return [ ...logs ].sort((a, b) => a.created < b.created ? -1 * sortDict.value : 1 * sortDict.value);
});

</script>

<style scoped>
.shadow-box {
    box-shadow: var(--el-box-shadow-lighter);
    margin-top: 1em;
}
</style>
