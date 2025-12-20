<template>
    <el-row
        :gutter="20"
        style="padding-top: 0.75em;">
        <el-col
            class="filter-col"
            :span="24"
            style="text-align:left;">
            <el-row :gutter="10">
                <el-col
                    :span="5"
                    style="text-align:center;">
                    <el-input
                        id="top-filters-text-search"
                        v-model="filters.search"
                        :style="{width:'100%'}"
                        placeholder="Текстовый поиск"
                        value-on-clear=""
                        clearable>
                        <template #prefix>
                            <el-icon class="el-input__icon">
                                <ElSearch />
                            </el-icon>
                        </template>
                    </el-input>
                </el-col>
                <el-col
                    :span="4"
                    style="text-align:center;">
                    <el-select
                        id="top-filters-istochnik"
                        v-model="filters.source"
                        :style="{width:'100%'}"
                        placeholder="Источник"
                        value-on-clear=""
                        clearable>
                        <el-option
                            v-for="item in scopeOptions.sources"
                            :key="item"
                            :label="item"
                            :value="item" />
                    </el-select>
                </el-col>
                <el-col
                    :span="4"
                    style="text-align:center;">
                    <el-select
                        id="top-filters-oblast"
                        :key="filters.scope"
                        v-model="filters.scope"
                        :style="{width:'100%'}"
                        placeholder="Область регулирования"
                        value-on-clear=""
                        clearable>
                        <el-option
                            v-for="item in scopeOptions.scopes"
                            :key="item"
                            :label="item"
                            :value="item" />
                    </el-select>
                </el-col>
                <el-col
                    :span="4"
                    style="text-align:center;">
                    <el-select
                        id="top-filters-lawtypes"
                        :key="filters.lawtype"
                        v-model="filters.lawtype"
                        :style="{width:'100%'}"
                        placeholder="Вид документа"
                        value-on-clear=""
                        clearable>
                        <el-option
                            v-for="item in scopeOptions.lawtypes"
                            :key="item"
                            :label="item"
                            :value="item" />
                    </el-select>
                </el-col>
                <el-col
                    :span="4"
                    style="text-align:center;">
                    <el-date-picker
                        v-model="filters.dates"
                        style="width:100%"
                        type="daterange"
                        format="DD.MM.YYYY"
                        range-separator=" - " />
                </el-col>
                <el-col
                    :span="3"
                    style="text-align:end;">
                    <el-button
                        id="menu-top-clear-filters"
                        type="danger"
                        @click="resetFilters">
                        Очистить фильтры
                    </el-button>
                </el-col>
            </el-row>
        </el-col>
    </el-row>
</template>

<script setup lang="ts">
import { Search as ElSearch } from "@element-plus/icons-vue";
import { type IFilters } from "@/api/models";

const {
    scopeOptions = {
        sources: [],
        scopes: [],
        lawtypes: [],
    },
} = defineProps<{
    scopeOptions?: {
        sources: string[];
        scopes: string[];
        lawtypes: string[];
    };
}>();

const filters = defineModel<IFilters>({ required: true });

const emit = defineEmits<{
    "reset-filters": [];
}>();
const resetFilters = () => {
    emit("reset-filters");
};

</script>
