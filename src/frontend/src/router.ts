import { createRouter, createWebHistory } from "vue-router";
import NotFoundPage from "@/pages/NotFoundPage.vue";
import MainPage from "@/pages/MainPage";
import DocumentCard from "@/pages/DocumentCard";
import RolesSettings from "@/pages/RolesSettings.vue";

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: "",
            alias: "/",
            component: MainPage,
        },
        {
            path: "/document",
            component: DocumentCard,
            props: url => ({ documentId: url.query.id }),
        },
        {
            path: "/roles_settings",
            component: RolesSettings,
        },
        {
            path: "/:catchAll(.*)",
            component: NotFoundPage,
        },
    ],
});

export default router;
