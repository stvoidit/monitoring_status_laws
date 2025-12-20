import locale from "dayjs/locale/ru";
import dayjs from "dayjs";
dayjs.locale(locale);
import App from "@/App.vue";
import { createApp } from "vue";
import { createPinia } from "pinia";
import router from "@/router";
createApp(App).use(router).use(createPinia()).mount("#app");
