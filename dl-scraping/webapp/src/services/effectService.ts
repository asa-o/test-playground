import { EffectInfo } from "../types/effects";
import { ResponseGetEffectList } from "../types/responses";

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
      pagerNextExists = responseData.isNext;
      page++;
      progressCallback(responseData.effects);
    }
  }

  renewalPossesionList(effects: EffectInfo[]): void {}
}

export default new EffectService();

///
// page nextがなくなるまで取得したデータはtempデータデーブルへ保持しておき、
// nextがなくなったタイミングで逆順にして正規のテーブルへ移動する
