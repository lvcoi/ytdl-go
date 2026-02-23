import { useNavigate } from '@solidjs/router';
import DownloadView from '../components/DownloadView';

export function Download() {
    const navigate = useNavigate();

    return (
        <DownloadView
            onOpenLibrary={() => navigate('/library')}
        />
    );
}
