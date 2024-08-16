import { EffectInfo } from "./effects";

export type ResponseGetEffectList = {
  dlSecKey: string;
  sessionId: string;
  effects: EffectInfo[];
  isNext: boolean;
};
