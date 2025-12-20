<template>
    <el-main>
        <el-row
            style="margin-bottom: 1em;"
            :gutter="20">
            <el-col
                :span="4"
                style="text-align:start">
                <el-button @click="$router.push('/')">
                    <el-icon>
                        <ElBack />
                    </el-icon>
                </el-button>
            </el-col>
            <el-col
                :offset="10"
                :span="6"
                style="text-align: center;">
                <el-select
                    v-model="filters.department"
                    filterable
                    clearable
                    placeholder="отдел">
                    <el-option
                        v-for="opt in uniqueDepartments"
                        :key="opt.value"
                        :value="opt.value"
                        :label="opt.text" />
                </el-select>
            </el-col>
            <el-col
                :span="2"
                style="text-align: center;">
                <el-checkbox
                    v-model="filters.isAdmin"
                    size="large">
                    Админ
                </el-checkbox>
            </el-col>
            <el-col
                :span="2"
                style="text-align: center;">
                <el-checkbox
                    v-model="filters.isResponsible"
                    size="large">
                    Ответственный
                </el-checkbox>
            </el-col>
        </el-row>

        <el-table
            v-loading="loadingUsers"
            :data="computedUsers"
            v-bind="tableProps">
            <el-table-column
                align="start"
                prop="department_label"
                label="Отдел" />
            <el-table-column
                align="start"
                label="ФИО"
                :formatter="userFormatter" />
            <el-table-column
                align="start"
                width="200"
                label="Роль">
                <template #default="{ row }: {row:MegaplanUser}">
                    <el-checkbox
                        :key="`${row.id}-${row.is_admin}`"
                        :model-value="row.is_admin"
                        :disabled="loading"
                        @change="(value) => onChange(row, 'is_admin', value)">
                        Админ
                    </el-checkbox>
                    <el-checkbox
                        :key="`${row.id}-${row.is_responsible}`"
                        :model-value="row.is_responsible"
                        :disabled="loading"
                        @change="(value) => onChange(row, 'is_responsible', value)">
                        Ответственный
                    </el-checkbox>
                </template>
            </el-table-column>
        </el-table>
    </el-main>
</template>

<script setup lang="ts">
import { type MegaplanUser } from "@/api/models";
import { computed, onMounted, reactive, h, shallowRef, ref, nextTick } from "vue";
import { ChangeRole } from "@/api";
import { ElText, ElSpace, type CheckboxValueType } from "element-plus";
import { MessageAlert, NotificationSaved } from "@/composibles/useAlert";
import { Back as ElBack } from "@element-plus/icons-vue";

const userFormatter = (row: MegaplanUser, _column, _cellValue) => {
    return h(ElSpace, {
        direction: "vertical",
        alignment: "start",
    }, {
        default: () => ([
            h(ElText, { style: { fontWeight: "bold" }, innerText: row.fio }),
            h(ElText, { innerText: row.position }),
            h(ElText, { innerText: `ID: ${row.id}` }),
        ]),
    });
};

import { createFetchLoading } from "@/composibles/useFetchLoading";

const users = shallowRef<MegaplanUser[]>([]);
const {
    loading: loadingUsers,
    doFetch: fetchMegaplanUser,
} = createFetchLoading("get", "/api/roles_settings", users);

onMounted(fetchMegaplanUser);

const loading = ref(false);
const onChange = async (user: MegaplanUser, prop: "is_admin" | "is_responsible", value: CheckboxValueType) => {
    loading.value = true;
    try {
        await ChangeRole({ ...user, [prop]: value });
        user[prop] = value as boolean;
        await nextTick();
        NotificationSaved();
    } catch (error) {
        MessageAlert((error as Error).message);
    } finally {
        loading.value = false;
    }
};
const uniqueDepartments = computed(() => {
    const likeSet = new Map<number, { text: string; value: number }>();
    for (const u of users.value) {
        if (!likeSet.has(u.department_id)) {
            likeSet.set(u.department_id, { text: u.department_label, value: u.department_id });
        }
    }
    return likeSet.values()
        .toArray()
        .sort((a, b) => b.text.localeCompare(a.text));
});

const filters = reactive({} as {
    department?: number;
    isAdmin?: boolean;
    isResponsible?: boolean;
});

const computedUsers = computed(() => {
    const filterDepartment = (user: MegaplanUser) => {
        if (!filters.department) return false;
        return user.department_id === filters.department;
    };
    const filterIsAdmin = (user: MegaplanUser) => {
        if (!filters.isAdmin) return false;
        return user.is_admin;
    };
    const filterIsExecutor = (user: MegaplanUser) => {
        if (!filters.isResponsible) return false;
        return user.is_responsible;
    };

    const makeFilterPipe = () => {
        const pipeList = [] as ((user: MegaplanUser) => boolean)[];
        if (filters.department) {
            pipeList.push(filterDepartment);
        }
        if (filters.isAdmin) {
            pipeList.push(filterIsAdmin);
        }
        if (filters.isResponsible) {
            pipeList.push(filterIsExecutor);
        }
        return (user: MegaplanUser) => {
            if (!pipeList.length) return true;
            return pipeList.some(fn => fn(user));
        };
    };
    const filtersPipeline = makeFilterPipe();
    return users.value.filter(filtersPipeline);
});

const tableProps = {
    rowKey: "id",
    height: "70vh",
    border: true,
    fit: true,
    flexible: true,
} as const;

</script>
