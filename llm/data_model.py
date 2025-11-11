from typing import List, Optional, Set, Tuple

from pydantic import BaseModel, Field


class Answer(BaseModel):
  buy: List[str] = Field(default_factory=list, description="list of softwares IDs to buy to achieve the story")
  rationale: str = Field("", description="explain the buying choice in less than 200 words")


class Story(BaseModel):
  identifier: int
  goal: str


class StoryImplementation(Story):
  failures: List[Answer] = Field(default_factory=list)
  success: Optional[Answer] = None
  strategy: str = ""
  owned: List[str] = Field(default_factory=list)
  left: List[str] = Field(default_factory=list)
  total_cost: int = 0


class StoryDescription(Story):
  available: Set[str] = Field(default_factory=set)
  accepted: List[Set[str]] = Field(default_factory=list)
  obsolete: Set[str] = Field(default_factory=set)
  tries: int = 5
  hints: List[Tuple[int, str]] = Field(default_factory=list)


class State(BaseModel):
  story: StoryDescription = None
  tentatively_owned: Set[str] = Field(default_factory=set)
  owned: Set[str] = Field(default_factory=set)
  left: Set[str] = Field(default_factory=set)
  past: List[StoryImplementation] = Field(default_factory=list)
  attempts: List[Answer] = Field(default_factory=list)
  total_cost: int = 0
  strategy: str = ""
  action: str = ""
  assessment: str = ""

