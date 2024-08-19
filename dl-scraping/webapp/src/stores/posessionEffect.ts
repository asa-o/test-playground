import { atom, createStore } from "jotai";
import { EffectInfo } from "@/types/effects";

export const posessionEffectAtom = atom<EffectInfo[]>([]);
export const posessionEffectStore = createStore();
