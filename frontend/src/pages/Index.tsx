import { useState, useEffect } from "react";
import { DeviceHeader } from "@/components/DeviceHeader";
import { MetricCard } from "@/components/MetricCard";
import { Zap, Activity, Gauge, Battery } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface DeviceStatus {
  code: string;
  value: boolean | number | string;
}

interface DeviceData {
  name: string;
  online: boolean;
  status: DeviceStatus[];
}

const DEBUG = import.meta.env.DEV;

const FALLBACK_DEVICE: DeviceData = {
  name: "Refrigerador",
  online: false,
  status: [
    { code: "switch_1", value: false },
    { code: "cur_power", value: 0 },
    { code: "cur_current", value: 0 },
    { code: "cur_voltage", value: 0 },
    { code: "add_ele", value: 0 },
  ],
};


const sanitizeTuya = (raw: any): DeviceData => {
  const result = raw?.result ?? {};
  const status: DeviceStatus[] = Array.isArray(result.status)
    ? result.status.filter((s: any) => s && typeof s.code === "string")
    : FALLBACK_DEVICE.status;

  return {
    name: typeof result.name === "string" && result.name.trim() ? result.name : FALLBACK_DEVICE.name,
    online: Boolean(result.online),
    status,
  };
};


function DebugPanel({ deviceData, lastUpdate }: { deviceData: any; lastUpdate: string }) {
  return (
    <details style={{ marginTop: 24, background: "rgba(0,0,0,.05)", padding: 12, borderRadius: 8 }}>
      <summary style={{ cursor: "pointer", fontWeight: 600 }}>Debug</summary>
      <div style={{ fontFamily: "monospace", whiteSpace: "pre-wrap", paddingTop: 8 }}>
        <div>lastUpdate: {lastUpdate}</div>
        <div>name: {deviceData?.name}</div>
        <div>online: {String(deviceData?.online)}</div>
        <div>status (ver também console.table)</div>
        <pre>{JSON.stringify(deviceData, null, 2)}</pre>
      </div>
    </details>
  );
}

const Index = () => {
  const [deviceData, setDeviceData] = useState<DeviceData | null>(null);
  const [lastUpdate, setLastUpdate] = useState<string>("");
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [showWelcome, setShowWelcome] = useState(true);
  const { toast } = useToast();


  useEffect(() => {
    if (DEBUG) {
      console.debug("[STATE] deviceData set:", deviceData);
      if (deviceData?.status) {
        console.table(deviceData.status.map((s) => ({ code: s.code, value: s.value })));
      }
    }
  }, [deviceData]);

  const fetchDeviceData = async () => {
    setIsRefreshing(true);

    const API = "/api";
    let updated = false;

    try {
      if (DEBUG) console.debug("[FETCH] iniciando:", `${API}/info`);
      const res = await fetch(`${API}/info`, { headers: { Accept: "application/json" } });

      const ct = res.headers.get("content-type") || "";
      if (DEBUG) console.debug("[FETCH] status:", res.status, "content-type:", ct);

      if (res.status === 429) {
        setDeviceData((prev) => prev ?? FALLBACK_DEVICE);
        updated = true;
        toast({
          title: "Informação não alterada",
          description: "Mostrando dados atuais. Tente novamente mais tarde.",
        });
      } else if (!res.ok) {
        if (DEBUG) console.error("[FETCH] erro HTTP:", res.status);
        setDeviceData(FALLBACK_DEVICE);
        updated = true;
        toast({
          title: "Erro ao buscar dados",
          description: `Código ${res.status}. Mostrando dados offline.`,
          variant: "destructive",
        });
      } else if (!ct.includes("application/json")) {
        const text = await res.text();
        if (DEBUG) console.error("[FETCH] resposta não-JSON, primeiros 300 chars:", text.slice(0, 300));
        setDeviceData(FALLBACK_DEVICE);
        updated = true;
        toast({
          title: "Resposta inesperada",
          description: "Mostrando dados offline.",
          variant: "destructive",
        });
      } else {
        const raw = await res.json();
        if (DEBUG) {
          console.debug("[FETCH] JSON cru:", raw);
          console.debug("[FETCH] JSON.result:", raw?.result);
        }
        const parsed = sanitizeTuya(raw);
        if (DEBUG) {
          console.debug("[PARSE] objeto normalizado:", parsed);
        }
        setDeviceData(parsed);
        updated = true;
        toast({
          title: "Dados atualizados",
          description: "Informações do dispositivo atualizadas com sucesso.",
        });
      }
    } catch (err) {
      if (DEBUG) console.error("[FETCH] erro de rede/parse:", err);
      setDeviceData(FALLBACK_DEVICE);
      updated = true;
      toast({
        title: "Erro de conexão",
        description: "Mostrando dados offline. Tente novamente.",
        variant: "destructive",
      });
    } finally {
      setLastUpdate(new Date().toLocaleString("pt-BR"));
      setShowWelcome(false);
      setIsRefreshing(false);

      if (!updated) {
        setDeviceData((prev) => prev ?? FALLBACK_DEVICE);
      }
    }
  };

  const getStatusValue = (code: string): number => {
    const status = deviceData?.status.find((s) => s.code === code);
    const v = typeof status?.value === "number" ? status.value : 0;
    if (DEBUG) console.debug("[USE] getStatusValue", code, "=>", v);
    return Number.isFinite(v) ? v : 0;
  };

  const getSwitchStatus = (): boolean => {
    const status = deviceData?.status.find((s) => s.code === "switch_1");
    const v = typeof status?.value === "boolean" ? status.value : false;
    if (DEBUG) console.debug("[USE] getSwitchStatus =>", v);
    return v;
  };

  if (showWelcome || !deviceData) {
    return (
      <div className="min-h-screen bg-gradient-primary dark flex items-center justify-center p-6">
        <Card className="max-w-2xl w-full bg-gradient-card border-border shadow-elegant">
          <CardHeader className="text-center space-y-4">
            <div className="mx-auto w-20 h-20 rounded-full bg-primary/20 flex items-center justify-center mb-4">
              <Gauge className="w-10 h-10 text-primary" />
            </div>
            <CardTitle className="text-4xl font-bold text-foreground">
              Monitor de Refrigeração
            </CardTitle>
            <p className="text-muted-foreground text-lg">
              Monitore em tempo real o consumo e status do seu aparelho de refrigeração através de smart plug
            </p>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="grid gap-4">
              <div className="flex items-start gap-3">
                <Zap className="w-5 h-5 text-primary mt-1" />
                <div>
                  <h3 className="font-semibold text-foreground">Potência em Tempo Real</h3>
                  <p className="text-sm text-muted-foreground">Acompanhe o consumo instantâneo em watts</p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <Activity className="w-5 h-5 text-primary mt-1" />
                <div>
                  <h3 className="font-semibold text-foreground">Medições Precisas</h3>
                  <p className="text-sm text-muted-foreground">Corrente, voltagem e consumo total</p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <Battery className="w-5 h-5 text-primary mt-1" />
                <div>
                  <h3 className="font-semibold text-foreground">Histórico de Consumo</h3>
                  <p className="text-sm text-muted-foreground">Visualize o consumo acumulado em kWh</p>
                </div>
              </div>
            </div>

            <button
              onClick={fetchDeviceData}
              disabled={isRefreshing}
              className="w-full bg-primary hover:bg-primary/90 text-primary-foreground font-semibold py-4 px-6 rounded-lg transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed shadow-elegant hover:shadow-glow"
            >
              {isRefreshing ? (
                <span className="flex items-center justify-center gap-2">
                  <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-primary-foreground"></div>
                  Conectando...
                </span>
              ) : (
                "Iniciar Monitoramento"
              )}
            </button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-primary dark p-6 md:p-8 lg:p-12">
      <div className="h-full max-w-[1800px] mx-auto flex flex-col gap-8">
        <DeviceHeader
          deviceName={deviceData.name}
          online={deviceData.online}
          lastUpdate={lastUpdate}
          onRefresh={fetchDeviceData}
          isRefreshing={isRefreshing}
        />

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 lg:gap-8">
          <MetricCard
            title="Potência Atual"
            value={getStatusValue("cur_power")}
            unit="W"
            icon={Zap}
          />
          <MetricCard
            title="Corrente"
            value={(getStatusValue("cur_current") / 1000).toFixed(2)}
            unit="A"
            icon={Activity}
          />
          <MetricCard
            title="Voltagem"
            value={(getStatusValue("cur_voltage") / 10).toFixed(1)}
            unit="V"
            icon={Gauge}
          />
          <MetricCard
            title="Consumo Total"
            value={(getStatusValue("add_ele") / 100).toFixed(2)}
            unit="kWh"
            icon={Battery}
          />
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 lg:gap-8 flex-1">
          <Card className="bg-gradient-card border-border flex flex-col">
            <CardHeader>
              <CardTitle className="text-foreground text-xl">Status do Aparelho</CardTitle>
            </CardHeader>
            <CardContent className="flex-1 flex items-center justify-center">
              <div className="flex flex-col items-center gap-6 py-8">
                <span className="text-muted-foreground text-lg">Estado atual:</span>
                <span className={`font-bold text-4xl ${getSwitchStatus() ? "text-success" : "text-destructive"}`}>
                  {getSwitchStatus() ? "LIGADO" : "DESLIGADO"}
                </span>
              </div>
            </CardContent>
          </Card>

          <Card className="bg-gradient-card border-border flex flex-col">
            <CardHeader>
              <CardTitle className="text-foreground text-xl">Informações do Dispositivo</CardTitle>
            </CardHeader>
            <CardContent className="flex-1 flex items-center">
              <div className="space-y-6 w-full">
                <div className="flex justify-between items-center py-3 border-b border-border/50">
                  <span className="text-muted-foreground text-lg">Modelo:</span>
                  <span className="text-foreground font-medium text-lg">{deviceData.name}</span>
                </div>
                <div className="flex justify-between items-center py-3 border-b border-border/50">
                  <span className="text-muted-foreground text-lg">Status:</span>
                  <span className="text-foreground font-mono text-2xl">{deviceData.online ? "🔵" : "🔴"}</span>
                </div>
                <div className="flex justify-between items-center py-3">
                  <span className="text-muted-foreground text-lg">Consumo adicional:</span>
                  <span className="text-foreground font-medium text-lg">{getStatusValue("add_ele")} Wh</span>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {DEBUG && <DebugPanel deviceData={deviceData} lastUpdate={lastUpdate} />}
      </div>
    </div>
  );
};

export default Index;
