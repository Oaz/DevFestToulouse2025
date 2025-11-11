import os
import sys
import json
import numpy as np
import pandas as pd
import plotly.express as px


def process_json_files(root_directory, max_score):
  """
  Process all JSON files in the directory hierarchy and create a pandas DataFrame
  """
  all_data = []

  # Walk through all directories and subdirectories
  for root, dirs, files in os.walk(root_directory):
    for file in files:
      if file.endswith('.json'):
        file_path = os.path.join(root, file)
        try:
          with open(file_path, 'r', encoding='utf-8') as f:
            data = json.load(f)

          # Extract basic information
          row = {
            'game': data.get('game', ''),
            'model': data.get('model', ''),
            'temperature': float(data.get('temperature', 0)),
            'start_time': pd.to_datetime(data.get('start', '')) if data.get('start') else None,
            'end_time': pd.to_datetime(data.get('end', '')) if data.get('end') else None,
            'total_cost': data.get('total_cost', None)
          }

          # Process history to extract story costs and asset ownership
          story_costs = {}
          asset_ownership = {}

          if 'history' in data:
            for story in data['history']:
              story_id = story.get('identifier')
              if story_id and 'success' in story:
                # Story was successful, record its cost
                story_costs[f'story_{story_id}'] = story.get('total_cost', 0)

                # Track assets owned at this story
                owned_assets = story.get('owned', [])
                for asset in owned_assets:
                  asset_name = f'asset_{asset}'
                  if asset_name not in asset_ownership:
                    asset_ownership[asset_name] = story_id

          # Add story costs to row
          row.update(story_costs)

          # Add asset ownership to row
          row.update(asset_ownership)

          all_data.append(row)

        except Exception as e:
          print(f"Error processing file {file_path}: {e}")
          continue

  # Create DataFrame
  df = pd.DataFrame(all_data)

  # Sort columns for better readability
  if not df.empty:
    # Get basic columns
    basic_cols = ['game', 'model', 'score', 'temperature', 'start_time', 'end_time', 'total_cost']
    df['score'] = df['total_cost'].fillna(max_score)

    # Get story columns and sort them
    story_cols = sorted([col for col in df.columns if col.startswith('story_')],
                        key=lambda x: int(x.split('_')[1]))

    # Get asset columns and sort them
    asset_cols = sorted([col for col in df.columns if col.startswith('asset_')])

    # Reorder columns
    column_order = basic_cols + story_cols + asset_cols
    df = df.reindex(columns=column_order)

  return df


def create_box_plot(df):
  """
  Create a Plotly box plot to visualize the distribution of scores for each model
  """
  if df.empty:
    print("No data available for plotting.")
    return

  # Sort models by name for consistent ordering
  sorted_models = sorted(df['model'].unique())

  # Apply gradient colors for models
  N = len(sorted_models)
  rainbow_colors = ['hsl('+str(h)+',50%'+',50%)' for h in np.linspace(0, 360, N)]
  color_map = dict(zip(sorted_models, rainbow_colors))
  df_plot = df.copy()
  df_plot['model_color'] = df_plot['model'].map(color_map)

  # Create the box plot with enhanced features
  fig = px.box(df_plot, x='model', y='score',
               title='Distribution of Scores by Model',
               labels={'model': 'Model', 'score': 'Score'},
               hover_data=['game', 'temperature'],
               category_orders={'model': sorted_models},  # Sort models by name
               points='all',  # Show all individual data points
               color='model',  # Color by model
               color_discrete_map=color_map)  # Use our gradient color mapping

  # Update traces to show mean and standard deviation
  fig.update_traces(boxmean='sd')  # Show mean and standard deviation

  # Update layout for better appearance
  fig.update_layout(
    xaxis_title="Model",
    yaxis_title="Score",
    showlegend=False,
    xaxis_tickangle=-45
  )
  
  # Save the plot as an HTML file
  fig.write_html("score_distribution_by_model.html")
  print("Box plot saved as 'score_distribution_by_model.html'")

  # Save the plot as an SVG file
  fig.write_image("score_distribution_by_model.svg")
  print("Box plot saved as 'score_distribution_by_model.svg'")

  # Optionally display the plot (uncomment the next line if you want to show it)
  fig.show()
  
  return fig


def export_games(df, asset_prefix, phase_delta):
  """
  Exports data for each game as a separate CSV file.
  """
  # Check if DataFrame is empty
  if df.empty:
    print("Cannot export CSV files as DataFrame is empty.")
    return

  cols = ["score"] + [col for col in df.columns if col.startswith("asset_")]

  # Iterate over unique games
  for game in df['game'].unique():
    # Select rows for the current game
    game_df = df[df['game'] == game][cols]

    # Save the data to a CSV file
    filename = f"{game}.csv"
    game_df.to_csv(filename, index=False)
    print(f"Saved game data for '{game}' to {filename}")

    # Extract relevant data
    game_data = []
    for row in game_df.itertuples():
      row_data = {'Assets': {}}
      for field in row._fields[1:]:
        value = getattr(row, field)
        field_name = field.replace("asset_", "")
        if not pd.notna(value):
          continue
        if field_name == "score":
          row_data['Score'] = int(value)
        else:
          row_data['Assets'][asset_prefix + field_name] = int(value) + phase_delta
      game_data.append(row_data)

    # Save the data to a JSON file
    filename = f"{game}.json"
    with open(filename, 'w', encoding='utf-8') as f:
      json.dump(game_data, f, indent=2)

    print(f"Saved game data for '{game}' to {filename}")


def main(max_score, asset_prefix, phase_delta):
  """
  Main function to process JSON files and save to CSV
  """
  # Set the root directory where JSON files are located
  # Change this to your actual directory path
  root_directory = '.'  # Current directory - change as needed

  print("Processing JSON files...")
  df = process_json_files(root_directory, max_score)
  
  export_games(df, asset_prefix, phase_delta)

  if df.empty:
    print("No JSON files found or processed successfully.")
    return

  print(f"Processed {len(df)} JSON files successfully.")
  print(f"DataFrame shape: {df.shape}")
  print("\nColumns found:")
  for col in df.columns:
    print(f"  - {col}")

  # Save to CSV
  output_file = 'game_data_summary.csv'
  df.to_csv(output_file, index=False)
  print(f"\nData saved to {output_file}")

  # Display first few rows for verification
  print("\nFirst few rows of the data:")
  print(df.head())

  # Show some statistics
  print(f"\nSummary statistics:")
  print(f"Games found: {df['game'].nunique()} unique games")
  print(f"Models found: {df['model'].nunique()} unique models")
  print(f"Temperature range: {df['temperature'].min():.1f} - {df['temperature'].max():.1f}")
  if 'total_cost' in df.columns:
    print(f"Total cost range: {df['total_cost'].min()} - {df['total_cost'].max()}")

  # Create and save the box plot
  print("\nCreating box plot...")
  create_box_plot(df)


if __name__ == "__main__":
  if len(sys.argv) < 2:
    print("Please provide game as command line argument")
    sys.exit(1)

  game = sys.argv[1]
  if game == "1":
    main(33, "R", 1)
  elif game == "2":
    main(114, "S", 9)
  else:
    print("Invalid game argument")