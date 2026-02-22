export const theme = {
  colors: {
    accent: {
      primary: '#8b5cf6', // Vibrant Purple
      secondary: '#3b82f6', // Electric Blue
      vibrant: '#f43f5e', // Rose/Red
    },
    gradients: {
      primary: 'linear-gradient(135deg, #8b5cf6 0%, #3b82f6 100%)',
      surface: 'linear-gradient(180deg, rgba(255, 255, 255, 0.05) 0%, rgba(255, 255, 255, 0) 100%)',
    }
  }
};

export const getStatusColor = (status) => {
  switch (status) {
    case 'completed': return 'text-green-400';
    case 'downloading': return 'text-blue-400';
    case 'error': return 'text-red-400';
    default: return 'text-gray-400';
  }
};
