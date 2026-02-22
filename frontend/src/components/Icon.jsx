import { Dynamic } from 'solid-js/web';
import {
  Zap, PlusCircle, Layers, Sliders, Puzzle, ChevronDown,
  Trash2, DownloadCloud, Music, ShieldCheck, Database,
  AlertCircle, CheckCircle2, Loader, Terminal,
  Search, Filter, Film, Play, ExternalLink,
  X, PlayCircle, SkipBack, Pause, SkipForward,
  LayoutDashboard, Move, GripVertical, Grid, Plus, Minus, Settings, BarChart2, HardDrive, Clock, Download, Edit2, Check, CornerRightDown,
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
  'move': Move,
  'grip': GripVertical,
  'grid': Grid,
  'plus': Plus,
  'minus': Minus,
  'settings': Settings,
  'bar-chart': BarChart2,
  'hard-drive': HardDrive,
  'clock': Clock,
  'download': Download,
  'edit-2': Edit2,
  'check': Check,
  'corner-right-down': CornerRightDown,
};

export default function Icon(props) {
  const component = iconMap[props.name];
  if (!component) {
    return <svg class={props.class} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/></svg>;
  }
  return <Dynamic component={component} class={props.class} />;
}
