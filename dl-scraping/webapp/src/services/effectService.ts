"use client";

import { EffectInfo } from "../types/effects";
import { useAtom } from "jotai";
import { posessionEffectAtom, posessionEffectStore } from "@/stores/posessionEffect";
import { ConnectInfoAtom, ConnectInfoStore } from "@/stores/workData";
import { ResponseGetEffectList } from "../types/responses";
import LocalDB from "@/stores/localDb";
import { connect } from "http2";

class EffectService {
  async getList(email: string, password: string, progressCallback: (progress: EffectInfo[]) => void): Promise<void> {
    let pagerNextExists = true;
    let sessionId = "";
    let page = 1;

    while (pagerNextExists) {
      const jsonData =
        sessionId === ""
          ? JSON.stringify({ mailAddress: email, password: password, page: page })
          : JSON.stringify({ sessionId: sessionId, page: page });

      const response = await fetch("http://localhost:8081/get-effect-list", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: jsonData,
      });
      const responseData: ResponseGetEffectList = await response.json();
      this.renewalPossesionList(responseData.effects);
      this.downloadImage(responseData.effects);
      pagerNextExists = responseData.isNext;
      page++;
      progressCallback(responseData.effects);

      ConnectInfoStore.set(ConnectInfoAtom, () => {
        return { dlSecKey: responseData.dlSecKey, sessionId: responseData.sessionId };
      });
    }
  }

  downloadImage(effects: EffectInfo[]): void {
    try {
      effects.map(async (effect) => {
        const response = await fetch("http://localhost:8081/get-effect-image", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ effectId: effect.Id }),
        });

        const data = await response.json();
        if (data.succeed) {
          const base64Response = await fetch(`data:image/jpeg;base64,${data.image}`);
          const blob = await base64Response.blob();
          LocalDB.getInstance().effectImage.add({ id: effect.Id, image: blob });
        }
      });
    } catch (e) {
      console.error(e);
    }
  }

  renewalPossesionList(effects: EffectInfo[]): void {
    posessionEffectStore.set(posessionEffectAtom, (prev) => {
      const newEffects = prev.concat(effects);
      return newEffects;
    });
  }

  async getImage(effectId: string): Promise<Blob | null> {
    const image = await LocalDB.getInstance().effectImage.get(effectId);
    return image ? image.image : null;
  }

  async getAllList() {}

  async getAllImages() {
    return LocalDB.getInstance().effectImage.getAllDatas();
  }

  async change(hdshId: string) {
    const connectInfo = ConnectInfoStore.get(ConnectInfoAtom);
    const response = await fetch("http://localhost:8081/change-effect", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ sessionId: connectInfo.sessionId, hashId: hdshId, dlSecKey: connectInfo.dlSecKey }),
    });

    const data = await response.json();
    if (data.succeed) {
      ConnectInfoStore.set(ConnectInfoAtom, () => {
        return { dlSecKey: data.dlSecKey, sessionId: data.sessionId };
      });
    }
  }
}

export default new EffectService();
