import { useCallback, useEffect, useState } from "react";
import {
  DeleteProvider,
  ListProviders,
  SaveProvider,
  SetActiveProvider,
  TestProvider,
} from "../../wailsjs/go/main/App";
import { app } from "../../wailsjs/go/models";
import type { ProviderProfile, TestResult } from "../types/provider";

type SaveInput = app.SaveInput;
// useProviders owns the settings list. Secrets never enter state:
// apiKey lives only in the form field until saved, then cleared.
export function useProviders() {
  const [providers, setProviders] = useState<ProviderProfile[]>([]);
  const [activeId, setActiveId] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [test, setTest] = useState<TestResult | null>(null);
  const [testing, setTesting] = useState(false);

  const refresh = useCallback(async () => {
    try {
      const list = (await ListProviders()) as ProviderProfile[];
      setProviders(list ?? []);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Gagal memuat provider.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const save = useCallback(
    async (in_: SaveInput) => {
      setError(null);
      try {
        const saved = (await SaveProvider(in_)) as ProviderProfile;
        if (!activeId) setActiveId(saved.id);
        await refresh();
        return saved;
      } catch (e) {
        setError(e instanceof Error ? e.message : "Gagal menyimpan.");
        return null;
      }
    },
    [activeId, refresh]
  );

  const remove = useCallback(
    async (id: string) => {
      setError(null);
      try {
        await DeleteProvider(id);
        if (activeId === id) setActiveId("");
        await refresh();
      } catch (e) {
        setError(e instanceof Error ? e.message : "Gagal menghapus.");
      }
    },
    [activeId, refresh]
  );

  const activate = useCallback(
    async (id: string) => {
      setError(null);
      try {
        await SetActiveProvider(id);
        setActiveId(id);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Gagal mengaktifkan.");
      }
    },
    []
  );

  const runTest = useCallback(async (in_: SaveInput) => {
    setTesting(true);
    setTest(null);
    try {
      const r = (await TestProvider(in_)) as TestResult;
      setTest(r);
      return r;
    } catch (e) {
      const r = { ok: false, message: e instanceof Error ? e.message : "Test gagal." };
      setTest(r);
      return r;
    } finally {
      setTesting(false);
    }
  }, []);

  return { providers, activeId, loading, error, test, testing, save, remove, activate, runTest, refresh };
}
