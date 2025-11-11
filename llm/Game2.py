from GameRunner import play
from Models.Anthropic import Anthropic
from Models.Cerebras import Cerebras
from Models.Google import Google
from Models.Ollama import Ollama
from Models.OpenAI import OpenAI

game = {
  "name": "game2",
  "context": """
  You are a system architect solving design challenges with programmable robots.
  You buy softwares to achieve a set of stories while minimizing the global cost of the solution.
  Software are provided as is and cannot be modified under any circumstances.
  Each robot can be programmed with only ONE software.
  Once a software is bought, it is available forever and it can be duplicated onto any robot at zero cost.
  Robots are unlimited in number and available for free. Only softwares have a cost.
  Robots can be arranged in any order at no cost.
  Robots can be named when the software uses names.
  Robots are hardwired to catch the ball when a ball is sent to them.
  You can complete any story by sending the ball to the robot of your choice and let the robots behave according to their software.
  For each story to solve, you select which software to buy (ONE or MORE), considering
    - software costs: lowest is preferred
    - software reuse: reusability lowers the total cost of the solution 
    - already owned softwares: you can buy only the ones you don't already own
  """,
  "example": """
  You are given a story as a target to achieve :
  Bob is in front of Alice. Program Alice so that the ball goes Alice→Bob.
  Available softwares to buy: ['A', 'B']

  The softwares descriptions are:
  [
    {"id": 'A', "cost": 3, "code": "When I receive the ball, I send it to Bob"},
    {"id": 'B', "cost": 5, "code": "When I receive the ball, I send it to the robot in front of me"}
  ]  
  
  You can either "buy":['A'] for a cost of 3 or "buy":['B'] for a cost of 5.
  Additionally, you give a short rationale to explain and remember the reason of the choice.
  
  if you think the next story asks "Program Bob so that the ball goes Alice→Bob→Alice."
  then it is better to immediately buy 'B' because this software can be used on both robots so that
  each of them send the ball to the robot in front of them, and you can solve both stories with a total cost of 5.

  If you think the next story asks something else then it might be better to buy 'A' because it is cheaper (3 against 5)
  and you have no immediate software reuse opportunity.  
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
    {"id": "C", "cost": 3, "code": "When I receive the ball, I throw it in the air"},
    {"id": "D", "cost": 7, "code": "When I receive the ball, I throw it up in the air and then send it to the robot that sent it to me"},
    {"id": "E", "cost": 7, "code": "When I receive the ball, I throw it up in the air and then send it to the robot on the right of me"},
    {"id": "F", "cost": 5, "code": "When I receive the ball, I throw it up in the air and then I throw it up in the air again"},
    {"id": "G", "cost": 5, "code": "When I receive the ball, I throw it up in the air and then bounce it"},
    {"id": "H", "cost": 7, "code": "When I receive the ball, I bounce it and then send it to the robot that sent it to me"},
    {"id": "I", "cost": 7, "code": "When I receive the ball, I bounce it and then send it to the robot on the right of me"},
    {"id": "J", "cost": 7, "code": "When I receive the ball, I throw it up in the air then I bounce it then I throw it up in the air"},
    {"id": "K", "cost": 3, "code": "When I hear 'Air', I send the ball to Alice"},
    {"id": "L", "cost": 9, "code": "When a robot shakes its arm, I send the ball to the robot that is next to the robot that shook its arm"},
    {"id": "M", "cost": 3, "code": "When I hear 'Air', I shake my arm"},
    {"id": "N", "cost": 7, "code": "When I hear 'Air', I send the ball to Alice and when I hear 'Bounce', I send the ball to Bob"},
    {"id": "O", "cost": 3, "code": "When I hear 'Bounce', I shake my arm"},
    {"id": "P", "cost": 11, "code": "When I hear 'Air', I send the ball to Alice and when I hear 'Bounce', I send the ball to Bob and when I hear 'Both', I send the ball to Claire"},
    {"id": "Q", "cost": 3, "code": "When I receive the ball, I send it to Eve"},
    {"id": "R", "cost": 3, "code": "When I hear 'Both' I shake my arm"},
    {"id": "S", "cost": 15, "code": "When I hear 'Air', I send the ball to Alice and when I hear 'Bounce', I send the ball to Bob and when I hear 'Both', I send the ball to Claire and when I hear 'Three', I send the ball to David"},
    {"id": "T", "cost": 3, "code": "When I hear 'Three', I shake my arm"},
  ],
  "stories": [
    {
      "id": 1,
      "goal": "The ball must go into the air.",
      "available": ["A", "B", "C", "D", "E"],
      "obsolete": [],
      "accepted": [["C"], ["D"], ["E"]],
      "tries": 4,
    },
    {
      "id": 2,
      "goal": "The ball must go into the air and then go into the air again and then stop.",
      "available": ["F"],
      "obsolete": ["C"],
      "accepted": [["E"], ["F"], ["D", "A"], ["D", "B"]],
      "tries": 4,
    },
    {
      "id": 3,
      "goal": "The ball must go in the air and then bounce and then stop.",
      "available": ["G", "H", "I"],
      "obsolete": ["F"],
      "accepted": [["G"], ["E", "I"]],
      "tries": 5,
    },
    {
      "id": 4,
      "goal": "The ball must go in the air and then bounce and then go in the air again and then stop.",
      "available": ["J"],
      "obsolete": ["G"],
      "accepted": [["J"], ["E", "I"]],
      "tries": 5,
    },
    {
      "id": 5,
      "goal": """
      Each time the user says 'Air', the ball must go into the air.
      The system of robots must loop and execute the expected action if the user says 'Air' multiple times.
      """,
      "available": ["K", "L", "M"],
      "obsolete": ["J"],
      "accepted": [["K", "D"], ["K", "B", "E"], ["L", "M", "D"], ["L", "M", "B", "E"], ["L", "M", "A", "E"]],
      "hints": [(1, "Think about correctly closing the loops")],
      "tries": 6,
    },
    {
      "id": 6,
      "goal": """
      Each time the user says 'Air', the ball must go into the air.
      Each time the user says 'Bounce', the ball must bounce.
      The system of robots must loop and execute the expected action if the user says 'Air' or 'Bounce' multiple times.
      """,
      "available": ["N", "O"],
      "obsolete": ["K"],
      "accepted": [
        ["N", "D", "H"],
        ["L", "M", "O", "D", "H"],
        ["L", "M", "O", "D", "B", "I"],
        ["L", "M", "O", "D", "A", "I"],
        ["L", "M", "O", "H", "B", "E"],
        ["L", "M", "O", "H", "A", "E"],
      ],
      "hints": [(1, "Think about correctly closing the loops")],
      "tries": 6,
    },
    {
      "id": 7,
      "goal": """
      Each time the user says 'Air', the ball must go into the air.
      Each time the user says 'Bounce', the ball must bounce.
      Each time the user says 'Both', the ball must go into the air and then bounce.
      The system of robots must loop and execute the expected action if the user says 'Air', 'Bounce', or 'Both' multiple times.
      """,
      "available": ["P", "Q", "R"],
      "obsolete": ["D", "H", "N"],
      "accepted": [
        ["P", "Q", "E", "I"],
        ["L", "M", "O", "R", "A", "E", "I"],
        ["L", "M", "O", "R", "B", "E", "I"],
      ],
      "hints": [(1, "Think about correctly closing the loops")],
      "tries": 8,
    },
    {
      "id": 8,
      "goal": """
      Each time the user says 'Air', the ball must go into the air.
      Each time the user says 'Bounce', the ball must bounce.
      Each time the user says 'Both', the ball must go into the air and then bounce.
      Each time the user says 'Three', the ball must go into the air and then bounce and then go in the air again.
      The system of robots must loop and execute the expected action if the user says 'Air', 'Bounce', 'Both', or 'Three' multiple times.
      """,
      "available": ["S", "T"],
      "obsolete": ["P"],
      "accepted": [
        ["S", "Q", "E", "I"],
        ["L", "M", "O", "R", "T", "A", "E", "I"],
        ["L", "M", "O", "R", "T", "B", "E", "I"],
      ],
      "hints": [(1, "Think about correctly closing the loops")],
      "tries": 8,
    },
  ]
}

models = [
  # [Ollama("llama3:8b"), True, 0],
  # [Ollama("llama3.1:8b"), True, 0],
  # [Ollama("codellama:7b"), True, 0],
  # [Ollama("dolphin-llama3:8b"), True, 0],
  # [Ollama("gemma3:4b"), True, 0],
  # [Ollama("codegemma:7b"), True, 0],
  # [Ollama("qwen2.5-coder:7b"), True, 0],
  # [Ollama("mistral:7b"), True, 0],
  # [Cerebras("llama-4-maverick-17b-128e-instruct"), True, 2],
  # [Cerebras("qwen-3-235b-a22b-instruct-2507"), True, 2],
  # [Google("gemini-2.5-flash"), True, 2],
  # [Google("gemini-2.5-pro"), True, 10],
  # [OpenAI("gpt-5"), True, 10],
  # [Anthropic("claude-sonnet-4-20250514"), True, 2],
  # [Anthropic("claude-sonnet-4-5-20250929"), True, 2],
  # [Anthropic("claude-opus-4-1-20250805"), True, 2],
]

for model, strategize, delay in models:
  play(
    game,
    model=model,
    strategize=strategize,
    delay=delay,
    temperatures=[
      (0, 3),
      (0.3, 4),
      (0.7, 4),
      (1, 4),
    ],
  )
