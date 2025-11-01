import { RefreshCw, Snowflake } from "lucide-react";
import { Button } from "@/components/ui/button";
import { StatusBadge } from "./StatusBadge";

interface DeviceHeaderProps {
  deviceName: string;
  online: boolean;
  lastUpdate: string;
  onRefresh: () => void;
  isRefreshing: boolean;
}

export const DeviceHeader = ({ 
  deviceName, 
  online, 
  lastUpdate, 
  onRefresh, 
  isRefreshing 
}: DeviceHeaderProps) => {
  return (
    <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-8">
      <div className="flex items-center gap-4">
        <div className="p-3 bg-gradient-accent rounded-lg">
          <Snowflake className="h-8 w-8 text-primary-foreground" />
        </div>
        <div>
          <h1 className="text-3xl font-bold text-foreground">Refrigerador</h1>
          <p className="text-sm text-muted-foreground">Última atualização: {lastUpdate}</p>
        </div>
      </div>
      <div className="flex items-center gap-4">
        <StatusBadge online={online} />
        <Button 
          onClick={onRefresh} 
          disabled={isRefreshing}
          className="bg-primary hover:bg-primary/90 text-primary-foreground shadow-glow"
        >
          <RefreshCw className={`mr-2 h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} />
          Atualizar
        </Button>
      </div>
    </div>
  );
};
