import pyvista as pv
import glob
import imageio.v2 as imageio

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
    for i in range(10):
        render_terrain(input_file=f"images/eroded_terrain{i}.ply", output_file=f"images/terrain_render_{i}.png")

    # Generate a GIF from the rendered images
    

    # Get all the rendered images in order
    image_files = sorted(glob.glob("images/terrain_render_*.png"))

    # Create a GIF with loop=0 (infinite looping)
    with imageio.get_writer("images/terrain_animation.gif", mode="I", loop=0) as writer:
        for image_file in image_files:
            image = imageio.imread(image_file)
            writer.append_data(image)

    print("GIF animation created at: images/terrain_animation.gif")