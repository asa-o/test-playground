import { atom } from "jotai";
import { ConnectInfo } from "@/types/connectInfo";

export const ConnectInfoAtom = atom<ConnectInfo>({ sessionId: "", dlSecKey: "" });
