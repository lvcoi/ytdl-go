import { Dynamic } from 'solid-js/web';
import {
  Zap, PlusCircle, Layers, Sliders, Puzzle, ChevronDown,
  Trash2, DownloadCloud, Music, ShieldCheck, Database,
  AlertCircle, CheckCircle2, Loader, Terminal,
  Search, Filter, Film, Play, ExternalLink,
  X, PlayCircle, SkipBack, Pause, SkipForward,
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
};

export default function Icon(props) {
  const component = iconMap[props.name] || iconMap['alert-circle'];
  return <Dynamic component={component} class={props.class} />;
}
