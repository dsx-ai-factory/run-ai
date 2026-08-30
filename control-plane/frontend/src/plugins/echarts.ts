import { use, registerTheme } from "echarts/core";
import { LineChart, BarChart, ScatterChart, CustomChart } from "echarts/charts";
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  MarkLineComponent,
  MarkAreaComponent,
  TitleComponent,
  ToolboxComponent,
  AriaComponent,
} from "echarts/components";
import { SVGRenderer } from "echarts/renderers";

use([
  LineChart,
  BarChart,
  ScatterChart,
  CustomChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  MarkLineComponent,
  MarkAreaComponent,
  TitleComponent,
  ToolboxComponent,
  // Enables ECharts `aria` (screen-reader descriptions + color-blind decals);
  // required for the aria option to take effect in the tree-shaken build.
  AriaComponent,
  SVGRenderer,
]);

registerTheme("runai", {
  textStyle: {
    fontFamily: "Roboto, sans-serif",
  },
});

export const RUNAI_ECHARTS_THEME_NAME = "runai";
