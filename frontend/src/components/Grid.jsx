import { splitProps } from 'solid-js';

export const Grid = (props) => {
  const [local, rest] = splitProps(props, ['children', 'class', 'cols']);
  
  // Default to a flexible media-first grid: 1 column on mobile, 2 on sm, 3 on md, 4 on lg, 5+ on xl
  const gridCols = local.cols || 'grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6';
  
  return (
    <div 
      class={`grid ${gridCols} gap-6 p-4 ${local.class || ''}`}
      {...rest}
    >
      {local.children}
    </div>
  );
};

export const GridItem = (props) => {
  const [local, rest] = splitProps(props, ['children', 'class']);
  
  return (
    <div 
      class={`glass rounded-2xl overflow-hidden transition-smooth hover:scale-[1.02] hover:shadow-vibrant ${local.class || ''}`}
      {...rest}
    >
      {local.children}
    </div>
  );
};
