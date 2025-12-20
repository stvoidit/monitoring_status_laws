<template>
    <el-menu
        style="margin-top: 1em;"
        :default-active="model"
        menu-trigger="click"
        :router="false">
        <el-menu-item
            v-for="item in filterItems"
            :id="item.propName"
            :key="item.propName"
            style="padding: 0 0.5em"
            :index="item.propName"
            @click="changeType">
            <el-icon :size="18">
                <component :is="item.icon" />
            </el-icon>
            <el-text
                style="width:100%"
                :type="item.type">
                {{ item.label }}
            </el-text>
            <el-tag
                effect="plain"
                style="min-width: 2.75em;"
                :disable-transitions="true"
                :type="item.type">
                {{ statusesCount[item.propName] ?? 0 }}
            </el-tag>
        </el-menu-item>
    </el-menu>
</template>

<script setup lang="ts">
import { type Component, h, reactive, watchEffect } from "vue";
import { type MenuItemRegistered } from "element-plus";
import {
    MessageBox as ElMessageBox,
    ChatDotSquare as ElChatDotSquare,
    CircleCheck as ElCircleCheck,
    Warning as ElWarning,
    Star as ElStar,
    CircleClose as ElCircleClose,
    Files as ElFiles,
    TakeawayBox as ElTakeawayBox,
} from "@element-plus/icons-vue";

const { statusesCount } = defineProps<{
    statusesCount: Record<string, number>;
}>();

const model = defineModel<string>({ default: "total" });
const changeType = (item: MenuItemRegistered) => {
    if (model.value !== item.index) {
        model.value = item.index;
    }
};

interface FilterItem {
    propName: string;
    label: string;
    icon: Component;
    type?: "primary";
}

const filterItems: FilterItem[] = reactive([
    {
        propName: "total",
        label: "Все законопроекты",
        icon: h(ElMessageBox),
    },
    {
        propName: "oncontrol",
        label: "На контроле",
        icon: h(ElChatDotSquare),
    },
    {
        propName: "completed",
        label: "Опубликованные",
        icon: h(ElCircleCheck),
    },
    {
        propName: "Important",
        label: "Важные",
        icon: h(ElWarning),
    },
    {
        propName: "favorites",
        label: "Избранные",
        icon: h(ElStar),
    },
    {
        propName: "cancelled",
        label: "Отклоненные",
        icon: h(ElCircleClose),
    },
    {
        propName: "draft",
        label: "Черновики",
        icon: h(ElFiles),
    },
    {
        propName: "archive",
        label: "Архив",
        icon: h(ElTakeawayBox),
    },
]);

watchEffect(() => {
    filterItems.forEach((item) => {
        if (model.value === item.propName) {
            item.type = "primary";
        } else {
            delete item.type;
        }
    });
}, { flush: "pre" });

</script>
<style>
.el-menu-item * {
    vertical-align: baseline;
}
</style>
