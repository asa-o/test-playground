import { openDB, IDBPDatabase } from "idb";
import { useState } from "react";

const DB_NAME = "my-database";

const EFFECT_IMAGE_STORE_NAME = "effect-image";

export class LocalDB {
  private static instance: LocalDB;
  private effectImageDB: LocalDBBase;

  private constructor() {
    this.effectImageDB = new LocalDBBase(EFFECT_IMAGE_STORE_NAME);
  }

  public static getInstance(): LocalDB {
    if (!LocalDB.instance) {
      LocalDB.instance = new LocalDB();
    }
    return LocalDB.instance;
  }

  get effectImage() {
    return this.effectImageDB;
  }
}

class LocalDBBase {
  private db: IDBPDatabase | null;
  private storeName: string;

  constructor(storeName: string) {
    this.storeName = storeName;
    this.db = null;
    this.initDB();
  }

  async initDB() {
    const storeName = this.storeName;
    const db = await openDB(DB_NAME, 1, {
      upgrade(db) {
        if (!db.objectStoreNames.contains(storeName)) {
          db.createObjectStore(storeName, { keyPath: "id", autoIncrement: true });
        }
      },
    });
    this.db = db;
  }

  async get(key: IDBValidKey) {
    return await this.db?.get(this.storeName, key);
  }

  async set(value: any) {
    return await this.db?.put(this.storeName, value);
  }

  async add(value: any) {
    const existing = await this.get(value.id);
    if (existing) {
      //      throw new Error(`ID ${value.id} already exists`);
      return await this.set(value);
    }
    return await this.db?.add(this.storeName, value);
  }

  async getAllDatas() {
    const tx = this.db?.transaction(this.storeName, "readonly");
    const store = tx?.objectStore(this.storeName);
    const allDatas = await store?.getAll();
    await tx?.done;
    return allDatas;
  }

  async delete(key: IDBValidKey) {
    return await this.db?.delete(this.storeName, key);
  }
}

export default LocalDB;
