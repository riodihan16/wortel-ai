import { useCallback, useEffect, useState } from "react";
import {
  ActiveProvider,
  DeleteProvider,
  ListProviders,
  SaveProvider,
  SetActiveModel,
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
  const [activeProvider, setActiveProvider] = useState<ProviderProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [test, setTest] = useState<TestResult | null>(null);
  const [testing, setTesting] = useState(false);

  const refresh = useCallback(async () => {
    try {
      const list = (await ListProviders()) as ProviderProfile[];
      setProviders(list ?? []);
      // ActiveProvider returns the profile directly, or null when none is
      // active (Wails drops a second non-error return value).
      const active = (await ActiveProvider()) as ProviderProfile | null;
      setActiveProvider(active ?? null);
      // The backend owns which provider is active; mirroring only local
      // clicks left every card showing "Aktifkan" after a reload.
      setActiveId(active?.id ?? "");
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
        await refresh();
        return saved;
      } catch (e) {
        setError(e instanceof Error ? e.message : "Gagal menyimpan.");
        return null;
      }
    },
    [refresh]
  );

  const remove = useCallback(
    async (id: string) => {
      setError(null);
      try {
        await DeleteProvider(id);
        await refresh();
      } catch (e) {
        setError(e instanceof Error ? e.message : "Gagal menghapus.");
      }
    },
    [refresh]
  );

  const activate = useCallback(
    async (id: string) => {
      setError(null);
      try {
        await SetActiveProvider(id);
        await refresh();
      } catch (e) {
        setError(e instanceof Error ? e.message : "Gagal mengaktifkan.");
      }
    },
    [refresh]
  );

  // setModel switches the chat model on the active provider, then refreshes so
  // the header picker and settings list stay consistent.
  const setModel = useCallback(
    async (id: string, model: string) => {
      setError(null);
      try {
        await SetActiveModel(id, model);
        await refresh();
      } catch (e) {
        setError(e instanceof Error ? e.message : "Gagal mengganti model.");
      }
    },
    [refresh]
  );

  const runTest = useCallback(async (in_: SaveInput) => {
    setTesting(true);
    setTest(null);
    try {
      const r = (await TestProvider(in_)) as TestResult;
      setTest(r);
      return r;
    } catch (e) {
      const r = { ok: false, message: e instanceof Error ? e.message : "Tes koneksi gagal." };
      setTest(r);
      return r;
    } finally {
      setTesting(false);
    }
  }, []);

  return { providers, activeId, activeProvider, loading, error, test, testing, save, remove, activate, setModel, runTest, refresh };
}
