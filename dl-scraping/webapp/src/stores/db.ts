import { openDB } from "idb";
import { useState, useEffect } from "react";

const DB_NAME = "my-database";
const STORE_NAME = "my-store";

const initDB = async () => {
  const db = await openDB(DB_NAME, 1, {
    upgrade(db) {
      if (!db.objectStoreNames.contains(STORE_NAME)) {
        db.createObjectStore(STORE_NAME, { keyPath: "id", autoIncrement: true });
      }
    },
  });
  return db;
};
