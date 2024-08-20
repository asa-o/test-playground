"use client";

import { useState } from "react";
import { useAtom } from "jotai";
import Image from "next/image";
import { useEffect } from "react";
import { EffectInfo } from "../types/effects";
import effectService from "@/services/effectService";
import LocalDB from "@/stores/localDb";
import styles from "./page.module.css";
import EffectListView from "@/components/EffectListView";
import { posessionEffectAtom, posessionEffectStore } from "@/stores/posessionEffect";

export default function Home() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [results, setResults] = useState<EffectInfo[]>([]);
  const [images, setImages] = useState<{ id: string; image: Blob }[]>([]);
  const posessionEffects = posessionEffectStore.get(posessionEffectAtom);
  const [count, setCount] = useState(0);

  useEffect(() => {
    // posessionEffectsの変更を監視
    console.log(posessionEffects);
  }, [posessionEffects]);

  const handleSubmit = async () => {
    try {
      await effectService.getList(email, password, (progress) => {
        progress.map((effect) => {
          console.log(effect.Name);
        });
        setCount((prev) => prev + progress.length);
      });
    } catch (e) {
      console.error(e);
    }
  };

  const handleGetAllImages = async () => {
    try {
      const imageResults = await effectService.getAllImages();
      if (imageResults != null) setImages(imageResults);
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
        <button onClick={handleGetAllImages}>画像を取得</button>
      </div>
      <div>
        {results.map((result, index) => (
          <div key={index}>
            <p>{result.Name}</p>
          </div>
        ))}
      </div>

      <p>取得数: {count}</p>
      {posessionEffects.length > 0 && (
        <>
          <EffectListView
            onImageClick={(item) => {
              effectService.change(item.HashId);
            }}
          />
          <div className={styles.imageContainer}>
            {images.map((image, index) => (
              <div key={index} className={styles.imageWrapper}>
                <p>ID: {image.id}</p>
                <img src={URL.createObjectURL(image.image)} alt={`Image ${index}`} className={styles.image} />
              </div>
            ))}
          </div>
        </>
      )}
    </>
  );
}
