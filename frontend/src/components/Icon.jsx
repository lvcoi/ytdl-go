import { Dynamic } from 'solid-js/web';
import {
  Zap, PlusCircle, Layers, Sliders, Puzzle, ChevronDown,
  Trash2, DownloadCloud, Music, ShieldCheck, Database,
  AlertCircle, CheckCircle2, Loader, Terminal,
  Search, Filter, Film, Play, ExternalLink,
  X, PlayCircle, SkipBack, Pause, SkipForward,
  LayoutDashboard, History, ChevronRight, ChevronLeft,
  Eye, EyeOff, Pencil, Check, Settings2, User, Maximize2,
  Home, BarChart2, Settings, Box, RefreshCw, RotateCcw, RotateCw, Plus, Info,
  Download, FileAudio, Video, WifiOff, AlertTriangle, Image,
  LayoutGrid, GripVertical, Scaling, Star,
} from 'lucide-solid';

const iconMap = {
  'zap': Zap,
  'plus-circle': PlusCircle,
  'layers': Layers,
  'sliders': Sliders,
  'puzzle': Puzzle,
  'chevron-down': ChevronDown,
  'trash-2': Trash2,
  'download-cloud': DownloadCloud,
  'music': Music,
  'shield-check': ShieldCheck,
  'database': Database,
  'alert-circle': AlertCircle,
  'check-circle-2': CheckCircle2,
  'loader': Loader,
  'terminal': Terminal,
  'search': Search,
  'filter': Filter,
  'film': Film,
  'play': Play,
  'external-link': ExternalLink,
  'x': X,
  'play-circle': PlayCircle,
  'skip-back': SkipBack,
  'pause': Pause,
  'skip-forward': SkipForward,
  'layout-dashboard': LayoutDashboard,
  'history': History,
  'chevron-right': ChevronRight,
  'chevron-left': ChevronLeft,
  'eye': Eye,
  'eye-off': EyeOff,
  'pencil': Pencil,
  'check': Check,
  'settings-2': Settings2,
  'user': User,
  'maximize-2': Maximize2,
  'home': Home,
  'bar-chart-2': BarChart2,
  'settings': Settings,
  'box': Box,
  'refresh-cw': RefreshCw,
  'rotate-ccw': RotateCcw,
  'rotate-cw': RotateCw,
  'plus': Plus,
  'info': Info,
  'download': Download,
  'file-audio': FileAudio,
  'video': Video,
  'wifi-off': WifiOff,
  'alert-triangle': AlertTriangle,
  'image': Image,
  'layout-grid': LayoutGrid,
  'grip': GripVertical,
  'scaling': Scaling,
  'star': Star,
};

export default function Icon(props) {
  const component = iconMap[props.name];
  if (!component) {
    if (import.meta.env.DEV) {
      console.warn(`[Icon] Unknown icon name: "${props.name}"`);
    }
    return <Dynamic component={iconMap['alert-circle']} class={props.class} />;
  }
  return <Dynamic component={component} class={props.class} />;
}
