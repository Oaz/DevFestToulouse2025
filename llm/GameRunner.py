from datetime import datetime
from time import sleep
from typing import Any, Tuple
import json
from langchain_core.language_models import BaseChatModel
from langchain_core.runnables import RunnableConfig
from langgraph.graph import StateGraph, START, END

from Models.Model import Model
from Prompt import Prompt
from data_model import *


def banner(text: str):
  print("=" * 50)
  print(text.center(50))
  print("=" * 50)


class GameRunner:
  def __init__(self, name: str, llm: BaseChatModel, temperature: float, strategize: bool, game, delay: int):
    self.name = name
    self.llm = llm
    self.temperature = temperature
    self.game = game
    self.delay = delay
    self.softwares = [
      {
        "id": s["id"],
        "cost": s["cost"],
        "code": s["code"],
      }
      for s in game["softwares"]
    ]
    self.stories = {
      s["id"]: StoryDescription(
        identifier=s['id'],
        goal=s['goal'],
        available=set(s["available"]),
        obsolete=set(s["obsolete"]),
        accepted=[set(accepted_group) for accepted_group in s["accepted"]],
        tries=s.get("tries", 5),
        hints=s.get("hints", []),
      ) for s in game["stories"]
    }
    self.story_graph = self.build_story_graph(strategize)
    self.completion_graph = self.build_completion_graph()
    self.prompt = self.build_prompt()
    self.saved_state = None

  def build_story_graph(self, strategize: bool):
    graph = StateGraph(State)
    if strategize:
      begin = "ask"
      graph.add_node("ask", self.ask)
      graph.add_edge("ask", "decision")
    else:
      begin = "decision"
    graph.add_edge(START, begin)
    graph.add_node("decision", self.decision)
    graph.add_edge("decision", END)
    return graph.compile(debug=True)

  def build_completion_graph(self):
    graph = StateGraph(State)
    graph.add_conditional_edges(
      START,
      self.is_game_complete,
      {
        "ongoing": "next",
        "complete": "conclusion",
      }
    )
    graph.add_node("next", self.next)
    graph.add_conditional_edges(
      "next",
      self.is_story_accepted,
      {
        "failure": END,
        "success": "skip",
      }
    )
    graph.add_node("skip", self.skip)
    graph.add_conditional_edges(
      "skip",
      self.is_game_complete,
      {
        "ongoing": "next",
        "complete": "conclusion",
      }
    )
    graph.add_node("conclusion", self.conclusion)
    graph.add_edge("conclusion", END)
    return graph.compile(debug=True)

  def build_prompt(self) -> Prompt:
    prompt = Prompt()
    prompt.add_static_section("context", self.game['context'])
    prompt.add_conditional_section("example",
                                   lambda state: state.story.identifier < 3 and len(state.strategy) == 0,
                                   self.game['example'])
    prompt.add_dynamic_section(
      "software_descriptions",
      lambda state: str([s for s in self.softwares if s['id'] in state.left or s['id'] in state.owned])
    )
    prompt.add_dynamic_section("current_story_to_solve", lambda state: state.story.goal)
    prompt.add_section("previously_bought_softwares", lambda state: len(state.owned) > 0,
                       lambda state: str(state.owned))
    prompt.add_dynamic_section("available_softwares", lambda state: str(state.left))
    prompt.add_conditional_section(
      "hint", lambda state: len(state.attempts) > 0,
      "You have previously attempted and failed to solve this story.")
    prompt.add_conditional_section(
      "hint", lambda state: len(state.attempts) > 0 and len(state.attempts[-1].buy) == 0,
      "You must always buy at least one software to solve the story. You cannot reply with an empty 'buy' list.")
    prompt.add_section(
      "hint", lambda state: len(state.attempts) > 2,
      lambda state:
      f"You have {len(state.attempts)} failed attempts on this story. Maybe it is time to act radically different."
    )
    prompt.add_section(
      "hint", lambda state: len(state.attempts) > 3,
      lambda state:
      "Acting radically different might just be buying more software, whatever the cost."
    )
    for story in self.stories.values():
      for hint in story.hints:
        prompt.add_conditional_section(
          "hint",
          lambda state, story_id=story.identifier, hint_level=hint[0]: state.story.identifier == story_id and len(
            state.attempts) > hint_level,
          hint[1]
        )
    prompt.add_section("failed_attempts", lambda state: len(state.attempts) > 0, lambda state: str(state.attempts[-4:]))
    prompt.add_section(
      "previously_completed_stories",
      lambda state: len(state.past) > 0,
      lambda state: str([{'goal': story.goal, 'success': story.success} for story in state.past])
    )
    prompt.add_section("strategy", lambda state: len(state.strategy) > 0, lambda state: state.strategy)
    prompt.add_dynamic_section("action", lambda state: state.action)
    return prompt

  def start(self) -> dict[str, Any]:
    self.saved_state = State(
      story=self.stories[1],
    )
    start_time = datetime.now()
    outcome = None
    complete = False
    while not complete:
      accepted_story = False
      tries = self.saved_state.story.tries
      while (not accepted_story) and (tries > 0):
        tries -= 1
        story_outcome = self.story_graph.invoke(self.saved_state)
        accepted_story = self.is_story_accepted(self.saved_state) == "success"
      if not accepted_story:
        self.record_story_implementation_failure(self.saved_state)
        return self.format_incomplete_result(self.saved_state, start_time)
      self.buy(self.saved_state)
      outcome = self.completion_graph.invoke(self.saved_state)
      complete = len(outcome['assessment']) > 0
    print("FINAL STATE", outcome)
    return self.format_result(outcome, start_time)

  def format_result(self, outcome: dict[str, Any], start_time: datetime) -> dict[str, Any]:
    return {
      'game': self.game['name'],
      'model': self.name,
      'temperature': self.temperature,
      'start': start_time.isoformat(),
      'end': datetime.now().isoformat(),
      'bought': list(outcome['owned']),
      'total_cost': outcome['total_cost'],
      'assessment': outcome['assessment'],
      'history': [self.format_story(story) for story in outcome['past']],
    }

  def format_incomplete_result(self, state: State, start_time: datetime) -> dict[str, Any]:
    return {
      'game': self.game['name'],
      'model': self.name,
      'temperature': self.temperature,
      'start': start_time.isoformat(),
      'end': datetime.now().isoformat(),
      'bought': list(state.owned),
      'history': [self.format_story(story) for story in state.past],
    }

  def format_story(self, story: StoryImplementation) -> dict[str, Any]:
    result = {
      'identifier': story.identifier,
      'goal': story.goal,
      'failures': [self.answer_result(answer) for answer in story.failures],
      'strategy': story.strategy,
      'total_cost': story.total_cost,
      'owned': list(story.owned),
      'left': list(story.left),
    }
    if story.success:
      result['success'] = self.answer_result(story.success)
    return result

  @staticmethod
  def answer_result(answer: Answer) -> dict[str, Any]:
    return {'buy': answer.buy, 'rationale': answer.rationale}

  def conclusion(self, state: State):
    self.record_story_implementation(state)
    prompt = f"""
    {self.game["conclusion"]}

    All existing softwares: {self.softwares}
    Softwares bought: {state.owned}
    Total cost: {state.total_cost}
    History: {state.past}
    """
    print("PROMPT", prompt)
    state.assessment = self.llm.invoke(prompt).content
    print("ASSESSMENT", state.assessment)
    sleep(self.delay)
    return state

  def ask(self, state: State):
    print('=====================================================================================')
    print("STATE", state)
    print("OWNED", state.owned)
    state.left = state.left.union(state.story.available).difference(state.story.obsolete).difference(state.owned)
    print("LEFT", state.left)
    state.strategy = ""
    state.action = """
     Think about a simple strategy. Do not over-complicate. 300 words max.
     List the softwares that would be good candidate to complete the story.
     """
    prompt_value = self.prompt.format(state)
    print("PROMPT", prompt_value)
    answer = self.llm.invoke(prompt_value)
    state.strategy = answer.content
    print("ANSWER", answer)
    sleep(self.delay)
    return state

  def decision(self, state: State):
    print('=====================================================================================')
    print("STATE", state)
    print("OWNED", state.owned)
    state.left = state.left.union(state.story.available).difference(state.story.obsolete).difference(state.owned)
    print("LEFT", state.left)
    state.action = """
    List the softwares that we should buy.
    Give a short rationale in less than 200 words.
    """
    prompt_value = self.prompt.format(state)
    print("PROMPT", prompt_value)
    answer: Answer = self.llm.with_structured_output(Answer).invoke(prompt_value)
    print("ANSWER", answer)
    sleep(self.delay)
    if not isinstance(answer, Answer):
      raise Exception("Should have been an Answer")
    answer.rationale = answer.rationale[:500]
    state.attempts.append(answer)
    state.tentatively_owned = state.owned.union(answer.buy)
    self.saved_state = state.model_copy(deep=True)
    return state

  @staticmethod
  def is_story_accepted(state: State):
    return "success" \
      if any(accepted_set.issubset(state.tentatively_owned) for accepted_set in state.story.accepted) \
      else "failure"

  def buy(self, state: State):
    print("BUY")
    state.owned = state.tentatively_owned
    state.total_cost = sum([software['cost'] for software in self.softwares if software['id'] in state.owned])
    state.left = state.left.difference(state.owned)
    self.saved_state = state.model_copy(deep=True)
    return state

  def is_game_complete(self, state: State):
    return "ongoing" if state.story.identifier < len(self.stories) else "complete"

  def next(self, state: State):
    print(f"NEXT STORY {state.story.identifier} to {state.story.identifier + 1}")
    self.record_story_implementation(state)
    state.attempts = []
    state.story = self.stories[state.story.identifier + 1]
    self.saved_state = state.model_copy(deep=True)
    return state.model_copy(deep=True)

  @staticmethod
  def record_story_implementation(state: State):
    state.past.append(StoryImplementation(
      identifier=state.story.identifier,
      goal=state.story.goal,
      failures=state.attempts[0:-1] if len(state.attempts) > 1 else [],
      success=state.attempts[-1] if len(state.attempts) > 0 else Answer(buy=[],
                                                                        rationale='Works with already owned softwares'),
      strategy=state.strategy,
      total_cost=state.total_cost,
      owned=list(state.owned),
      left=list(state.left),
    ))

  @staticmethod
  def record_story_implementation_failure(state: State):
    state.past.append(StoryImplementation(
      identifier=state.story.identifier,
      goal=state.story.goal,
      failures=state.attempts,
      success=None,
      strategy=state.strategy,
      total_cost=state.total_cost,
      owned=list(state.owned),
      left=list(state.left),
    ))

  @staticmethod
  def skip(state: State):
    print(f"SKIP {state.story.identifier}")
    return state


def play(game: dict, model: Model, strategize: bool, temperatures: List[Tuple[float, int]], delay: int = 0):
  banner(model.name)
  for temperature, runs in temperatures:
    llm = model.create(temperature=temperature)
    for run in range(runs):
      runner = GameRunner(model.name, llm, temperature, strategize, game, delay)
      result = runner.start()
      json_result = json.dumps(result, indent=2)
      filename = f"{game['name']}_{model.name}_{temperature}_{result['start']}.json"
      with open(f"results/{filename}", "w") as f:
        f.write(json_result)
      print(json_result)
