export const featureIconKeys = ["broom", "eye", "terminal", "bolt", "shield", "code"] as const;
export type FeatureIcon = (typeof featureIconKeys)[number];

export interface Feature {
  icon: FeatureIcon;
  title: string;
  desc: string;
}

export interface StepCard {
  step: string;
  stepColor: "accent" | "amber";
  title: string;
  desc: string;
  code?: string;
}

export type ComparisonVariant = "Manual Commands" | "Generic Cleaners" | "Clean Wizard";

export interface ComparisonItem {
  variant: ComparisonVariant;
  pros: string[];
  cons: string[];
  accent: boolean;
}

export type MatrixValue = "yes" | "no" | "partial" | string;

export interface MatrixRow {
  feature: string;
  values: [MatrixValue, MatrixValue, MatrixValue];
}

export interface ComparisonMatrix {
  columns: [ComparisonVariant, ComparisonVariant, ComparisonVariant];
  rows: MatrixRow[];
}

export const useCaseIconKeys = ["dev", "server", "ci", "bolt", "check"] as const;
export type UseCaseIcon = (typeof useCaseIconKeys)[number];

export interface UseCase {
  title: string;
  desc: string;
  icon: UseCaseIcon;
}

export interface PainPoint {
  tool: string;
  detail: string;
  size: string;
}

export const uiIconKeys = [
  "arrow-external",
  "arrow-right",
  "github",
  "menu",
  "close",
  "sun",
  "moon",
  "star",
] as const;
export type UIIcon = (typeof uiIconKeys)[number];

export type IconName = FeatureIcon | UseCaseIcon | UIIcon;
