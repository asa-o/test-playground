import React from "react";
import { useAtomValue } from "jotai";
import { FixedSizeList as List } from "react-window";
import { EffectInfo } from "@/types/effects";
import effectService from "@/services/effectService";
import { posessionEffectAtom, posessionEffectStore } from "@/stores/posessionEffect";
import styles from "./EffectListView.module.css";

interface ItemProps {
  index: number;
  style: React.CSSProperties;
  data: EffectInfo[];
}

const Row: React.FC<ItemProps> = ({ index, style, data }) => {
  const [imageSrc, setImageSrc] = React.useState<string[]>([]);
  const itemsPerRow = 3;
  const startIndex = index * itemsPerRow;
  const items = data.slice(startIndex, startIndex + itemsPerRow);

  React.useEffect(() => {
    const fetchImage = async () => {
      items.map(async (item, index) => {
        const imageBlob = await effectService.getImage(item.Id);
        if (imageBlob) {
          setImageSrc((prev) => {
            if (prev != null) {
              prev[index] = URL.createObjectURL(imageBlob);
            }
            return prev;
          });
        }
      });
    };
    fetchImage();
  }, [data, index]);

  return (
    <div className={styles.row} style={style}>
      {items.map((item, i) => (
        <div key={i} className={styles.item}>
          {item.Name}
          {imageSrc && imageSrc[i] && (
            <img src={imageSrc[i]} alt={`Image ${startIndex + i}`} className={styles.image} />
          )}
        </div>
      ))}
      {/* 空のdivを追加して、3つのアイテムが揃うようにする */}
      {Array.from({ length: itemsPerRow - items.length }).map((_, i) => (
        <div key={`empty-${i}`} className={styles.item}></div>
      ))}
    </div>
  );
};

const EffectListView: React.FC = () => {
  const posessionEffects = posessionEffectStore.get(posessionEffectAtom);
  const itemCount = Math.ceil(posessionEffects.length / 3);

  return (
    <List height={1000} itemCount={itemCount} itemSize={300} width={1200} itemData={posessionEffects}>
      {Row}
    </List>
  );
};

export default EffectListView;
