"use client";

import { EffectInfo } from "../types/effects";
import { ResponseGetEffectList } from "../types/responses";
import LocalDB from "@/stores/localDb";

const sprintf = require("sprintf-js").sprintf;

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

  renewalPossesionList(effects: EffectInfo[]): void {}

  async getAllImages() {
    return LocalDB.getInstance().effectImage.getAllDatas();
  }
}

export default new EffectService();
