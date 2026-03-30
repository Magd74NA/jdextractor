export interface ChartColors {
  primary: string;
  text: string;
  muted: string;
  border: string;
  background: string;
  success: string;
  danger: string;
}

function getCSSVar(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
}

export function getChartColors(): ChartColors {
  return {
    primary: getCSSVar('--pico-primary'),
    text: getCSSVar('--pico-color'),
    muted: getCSSVar('--pico-muted-color'),
    border: getCSSVar('--pico-muted-border-color'),
    background: getCSSVar('--pico-background-color'),
    success: getCSSVar('--pico-ins-color'),
    danger: getCSSVar('--pico-del-color'),
  };
}
