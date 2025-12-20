import {
    Top as ElTop,
    Bottom as ElBottom,
} from "@element-plus/icons-vue";
import { ref, computed, markRaw } from "vue";

export default () => {
    const sortDict = ref(-1);
    const handleChangeDirection = () => {
        sortDict.value *= -1;
    };
    const directionIcon = computed(() => {
        return sortDict.value < 0 ? markRaw(ElTop) : markRaw(ElBottom);
    });
    return {
        sortDict,
        directionIcon,
        handleChangeDirection,
    };
};
