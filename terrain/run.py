#!/usr/bin/env python3
import pyvista as pv
import glob
import imageio.v2 as imageio
import argparse
import os

def render_terrain(input_file="images/eroded_terrain.ply", 
                   output_file="terrain_render.png",
                   cmap="terrain",
                   background_color="black",
                   view_type="isometric"):
    """
    Renderiza un archivo de terreno 3D y guarda la imagen resultante.
    
    Args:
        input_file: Ruta al archivo .ply de entrada
        output_file: Ruta donde guardar la imagen renderizada
        cmap: Mapa de colores para la visualización
        background_color: Color de fondo para el render
        view_type: Tipo de vista ('isometric', 'xy', 'xz', 'yz')
    
    Returns:
        La ruta del archivo de imagen generado
    """
    # Cargar el archivo .ply
    mesh = pv.read(input_file)

    # Calcular escalares de elevación si es necesario
    if not mesh.point_data:
        mesh = mesh.elevation()

    # Inicializar plotter con fondo transparente
    plotter = pv.Plotter(off_screen=True)
    plotter.set_background(color=background_color)

    # Añadir malla con mapa de colores
    plotter.add_mesh(
        mesh,
        cmap=cmap,
        show_edges=False,
        scalars=mesh.active_scalars_name,
        opacity=1.0,
        lighting=True,
        show_scalar_bar=False,
    )

    # Configurar vista y iluminación
    if view_type == "isometric":
        plotter.view_isometric()
    elif view_type == "xy":
        plotter.view_xy()
    elif view_type == "xz":
        plotter.view_xz()
    elif view_type == "yz":
        plotter.view_yz()
    
    plotter.set_viewup([0, 0, 1])

    # Renderizar y guardar la imagen
    plotter.show(screenshot=output_file)
    
    return output_file

if __name__ == "__main__":
    # Configurar parser de argumentos
    parser = argparse.ArgumentParser(description="Renderizador de terreno 3D")
    parser.add_argument("-i", "--input", default="images/eroded_terrain.ply",
                        help="Ruta al archivo .ply de entrada")
    parser.add_argument("-o", "--output", default="terrain_render.png",
                        help="Ruta donde guardar la imagen renderizada")
    parser.add_argument("-c", "--cmap", default="terrain",
                        help="Mapa de colores para la visualización")
    parser.add_argument("-b", "--background", default="black",
                        help="Color de fondo para el render")
    parser.add_argument("-v", "--view", default="isometric",
                        choices=["isometric", "xy", "xz", "yz"],
                        help="Tipo de vista")
    
    args = parser.parse_args()
    
    # Verificar que el archivo de entrada existe
    if not os.path.exists(args.input):
        print(f"ERROR: El archivo {args.input} no existe")
        exit(1)
    
    # Llamar a la función con los parámetros de la línea de comandos
    output = render_terrain(
        input_file=args.input,
        output_file=args.output,
        cmap=args.cmap,
        background_color=args.background,
        view_type=args.view
    )
    
    print(f"Imagen generada correctamente: {output}")
