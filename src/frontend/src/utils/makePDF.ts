import { createPdf } from "pdfmake";
import normalFont from "@/assets/fonts/TT Moscow Economy Normal.ttf?url";
import boldFont from "@/assets/fonts/TT Moscow Economy Bold.ttf?url";
import italicsFont from "@/assets/fonts/TT Moscow Economy Italic.ttf?url";
import bolditalicsFont from "@/assets/fonts/TT Moscow Economy Bold Italic.ttf?url";
const fonts = {
    ["TT Moscow Economy"]: {
        normal: `${location.origin}${normalFont}`,
        bold: `${location.origin}${boldFont}`,
        italics: `${location.origin}${italicsFont}`,
        bolditalics: `${location.origin}${bolditalicsFont}`,
    },
};

export function makePDF(body: { headers: unknown[]; data: unknown[] }) {
    body.data.unshift(body.headers);
    const docDefinition = {
        pageSize: "A4",
        pageOrientation: "landscape",
        defaultStyle: {
            font: "TT Moscow Economy",
            fontSize: 10,
            bold: false,
        },
        content: [
            {
                table: {
                    headerRows: 1,
                    widths: body.headers.map(() => "auto"),
                    body: body.data,
                },
            },
        ],
    };
    createPdf(docDefinition, {}, fonts).download("документы.pdf");
}
