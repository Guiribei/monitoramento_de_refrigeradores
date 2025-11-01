import { Badge } from "@/components/ui/badge";
import { Wifi, WifiOff } from "lucide-react";

interface StatusBadgeProps {
  online: boolean;
}

export const StatusBadge = ({ online }: StatusBadgeProps) => {
  return (
    <Badge 
      variant={online ? "default" : "destructive"}
      className="flex items-center gap-2 px-4 py-2 text-sm font-medium"
    >
      {online ? (
        <>
          <Wifi className="h-4 w-4" />
          Online
        </>
      ) : (
        <>
          <WifiOff className="h-4 w-4" />
          Offline
        </>
      )}
    </Badge>
  );
};
