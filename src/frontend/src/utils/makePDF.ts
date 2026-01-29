import normalFont from "@/assets/fonts/TT Moscow Economy Normal.ttf?url";
import boldFont from "@/assets/fonts/TT Moscow Economy Bold.ttf?url";
import italicsFont from "@/assets/fonts/TT Moscow Economy Italic.ttf?url";
import bolditalicsFont from "@/assets/fonts/TT Moscow Economy Bold Italic.ttf?url";
import type { LawDocument } from "@/api/models";

export async function makePDF(body: { headers: unknown[]; data: unknown[] }) {
    const fonts = {
        ["TT Moscow Economy"]: {
            normal: `${location.origin}${normalFont}`,
            bold: `${location.origin}${boldFont}`,
            italics: `${location.origin}${italicsFont}`,
            bolditalics: `${location.origin}${bolditalicsFont}`,
        },
    };
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
    const { default: pdfMake } = await import("pdfmake");
    pdfMake.setFonts(fonts);
    pdfMake.createPdf(docDefinition).download("документы.pdf");
}

export const fnMapData = (visibleProps: string[]) => {
    return (d: LawDocument, index: number) => d.ArrayValues(index + 1, visibleProps);
};

const visibleFields = [
    { label: "#", prop: "#" },
    { label: "Краткое название", prop: "short_label" },
    { label: "Краткое содержание", prop: "desc" },
    { label: "Примечания", prop: "note" },
    { label: "Название", prop: "label" },
    { label: "Статус", prop: "status" },
    { label: "Идентификатор", prop: "project" },
];
const visibleHeaders = visibleFields.map(f => f.label);
const visibleProps = visibleFields.map(f => f.prop);

export const downloadPDF = async (tableData: LawDocument[]) => {
    const contentData = tableData.map(fnMapData(visibleProps));
    const buildData = {
        headers: visibleHeaders,
        data: contentData,
    };
    await makePDF(buildData);
};
