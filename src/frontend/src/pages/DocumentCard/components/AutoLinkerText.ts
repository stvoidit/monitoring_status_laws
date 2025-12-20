import { ElText } from "element-plus";
import { h } from "vue";
import Autolinker from "autolinker";
import dayjs from "dayjs";

const AutoLinkerText = (props: { text: string | number | Date | null }) => {
    let innerHTML = "";
    if (props.text instanceof Date) {
        innerHTML = dayjs(props.text).format("DD.MM.YYYY");
    }
    if (props.text && Number.isInteger(props.text)) {
        innerHTML = props.text.toString();
    }
    if (typeof props.text === "string") {
        innerHTML = Autolinker.link(props.text, {
            stripPrefix: false,
            stripTrailingSlash: false,
        });
    }
    return h(ElText, { innerHTML });
};

export default AutoLinkerText;
