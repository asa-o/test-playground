"use client";

import { useState } from "react";
import { useAtom } from "jotai";
import { posessionEffectAtom } from "@/stores/posessionEffect";
import Image from "next/image";
import { useEffect } from "react";
import { EffectInfo } from "../types/effects";
import effectService from "@/services/effectService";

export default function Home() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [results, setResults] = useState<EffectInfo[]>([]);

  useEffect(() => {}, []);

  const handleSubmit = async () => {
    try {
      await effectService.getList(email, password, (progress) => {
        progress.map((effect) => {
          console.log(effect.Name);
        });
      });
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <>
      <div>
        <input type="email" placeholder="メールアドレス" value={email} onChange={(e) => setEmail(e.target.value)} />
        <input
          type="password"
          placeholder="パスワード"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <button onClick={handleSubmit}>実行</button>
      </div>
      <div>
        {results.map((result, index) => (
          <div key={index}>
            <p>{result.Name}</p>
          </div>
        ))}
      </div>
    </>
  );
}
