from GameRunner import play
from Models.Anthropic import Anthropic
from Models.Cerebras import Cerebras
from Models.Google import Google
from Models.Ollama import Ollama
from Models.OpenAI import OpenAI

game = {
  "name": "game1",
  "context": """
  You are a system architect solving design challenges with programmable robots.
  You buy softwares to achieve a set of stories while minimizing the global cost of the solution.
  Each robot has a name and can be programmed with only 1 software.
  Once a software is bought, it is available forever and it can be duplicated on any robot at zero cost.
  For each story to solve, you select which software to buy, considering software costs, software reuse and already owned softwares.
  """,
  "example": """
  The softwares descriptions are:
  [{"id": 'A', "cost": 3, "code": "When I receive the ball, I send it to Bob"},{"id": 'B', "cost": 5, "code": "When I receive the ball, I send it to the robot in front of me"}]
  
  You are given a story as a target to achieve :
  Bob is in front of Alice. Program Alice so that the ball goes Alice→Bob.
  Available softwares to buy: ['A', 'B']
  
  You can either "buy":['A'] for a cost of 3 or "buy":['B'] for a cost of 5.
  Additionally, you give a rationale to explain and remember the reason of the choice.
  
  If the next story asks "Claire is in on the left of Bob. Program Bob so that the ball goes Alice→Bob→Claire."
  then it would have been better to buy 'A' because it is cheaper and neither 'A' or 'B' could solve this next story.
  
  if the next story asks "Program Bob so that the ball goes Alice→Bob→Alice."
  then it would have been better to buy 'B' because it would have also solved this next story without buying anything else.
  """,
  "conclusion": """
    We are doing a post-mortem evaluation system architecture decisions on programmable robots project.
    The evaluation audience is C-level executives so you have to keep it short: less than 500 words.
    Stay factual. Write in a neutral tone. Do not show any sentiment.
    Do not use meaningless descriptions such as "some", "key", "foundation" or any other shallow vocabulary 
    Do not question the story goals. Focus on the buying decisions.
    List software which were not bought and could have been a better buying decision.
  """,
  "softwares": [
    {"id": "A", "cost": 3, "code": "When I receive the ball, I send it to Alice"},
    {"id": "B", "cost": 3, "code": "When I receive the ball, I send it to Bob"},
    {"id": "C", "cost": 5, "code": "When I receive the ball, I send it to the robot in front of me"},
    {"id": "D", "cost": 3, "code": "When I receive the ball, I send it to Claire"},
    {"id": "E", "cost": 5, "code": "When I receive the ball, I send it to the robot on the right of me"},
    {"id": "F", "cost": 3, "code": "When I receive the ball, I send it to David"},
    {"id": "G", "cost": 5, "code": "When I receive the ball, I send it to the robot on the left of me"},
    {"id": "H", "cost": 3, "code": "When I receive the ball, I send it to Eve"},
    {"id": "I", "cost": 3, "code": "When I receive the ball, I send it to Fred"},
  ],
  "stories": [
    {
      "id": 1,
      "goal": "Bob is in front of Alice. Program Alice so that the ball goes Alice→Bob.",
      "available": ["B", "C"],
      "obsolete": [],
      "accepted": [["B"], ["C"]]
    },
    {
      "id": 2,
      "goal": "Claire is on the right of Bob. Program Bob so that the ball goes Alice→Bob→Claire.",
      "available": ["D", "E"],
      "obsolete": ["B", "C"],
      "accepted": [["D"], ["E"]]
    },
    {
      "id": 3,
      "goal": "David is on the left of Claire. Program Claire so that the ball goes Alice→Bob→Claire→David.",
      "available": ["F", "G"],
      "obsolete": ["D"],
      "accepted": [["F"], ["G"]]
    },
    {
      "id": 4,
      "goal": "Eve is on the right of David. Program David so that the ball goes Alice→Bob→Claire→David→Eve.",
      "available": ["H"],
      "obsolete": ["F"],
      "accepted": [["H"], ["E"]]
    },
    {
      "id": 5,
      "goal": "Fred is on the right of Eve. Program Eve so that the ball goes Alice→Bob→Claire→David→Eve→Fred.",
      "available": ["I"],
      "obsolete": ["H"],
      "accepted": [["I"], ["E"]]
    },
    {
      "id": 6,
      "goal": "Alice is on the right of Fred. Program Fred so that the ball goes Alice→Bob→Claire→David→Eve→Fred→Alice.",
      "available": ["A"],
      "obsolete": ["I"],
      "accepted": [["A"], ["E"]]
    },
  ]
}

models = [
  # [Ollama("llama3:8b"), 0],
  # [Ollama("llama3.1:8b"), 0],
  # [Ollama("codellama:7b"), 0],
  # [Ollama("dolphin-llama3:8b"), 0],
  # [Ollama("gemma3:4b"), 0],
  # [Ollama("codegemma:7b"), 0],
  # [Ollama("qwen2.5-coder:7b"), 0],
  # [Ollama("mistral:7b"), 0],
  # [Ollama("mistral-small:24b"), 0],
  # [Ollama("mistral-nemo:12b"), 0],
  # [Ollama("openthinker:7b"), 0],
  # [Cerebras("llama-4-maverick-17b-128e-instruct"), 2],
  # [Cerebras("qwen-3-235b-a22b-instruct-2507"), 2],
  # [Google("gemini-2.5-flash"), 2],
  # [Google("gemini-2.5-pro"), 5],
  # [OpenAI("gpt-5"), 10],
  # [Anthropic("claude-sonnet-4-20250514"), 2],
  # [Anthropic("claude-sonnet-4-5-20250929"), 2],
  # [Anthropic("claude-opus-4-1-20250805"), 2],
]

for model, delay in models:
  play(
    game=game,
    model=model,
    strategize=False,
    delay=delay,
    temperatures=[
      (0, 3),
      (0.3, 4),
      (0.7, 4),
      (1, 4),
    ],
  )
