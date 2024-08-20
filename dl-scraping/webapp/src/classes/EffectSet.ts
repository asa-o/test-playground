import { EffectInfo } from "../types/effects";

class EffectSet {
  private effects: EffectInfo[] = [];

  constructor() {}

  async fetchEffectList(email: string, password: string): Promise<void> {
    try {
      const jsonData = JSON.stringify({ sessionId: "", page: 1, mailAddress: email, password: password });
      const response = await fetch("https://asia-northeast1-asa-o-experiment.cloudfunctions.net/get-effect-list", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        mode: "cors",
        body: jsonData,
      });
      const data = await response.json();
      this.effects = data.effects;
    } catch (e) {
      console.error(e);
    }
  }

  getEffects(): EffectInfo[] {
    return this.effects;
  }
}

export default EffectSet;
