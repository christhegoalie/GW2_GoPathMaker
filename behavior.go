package main

type Behavior int

const (
	Behavior_Default Behavior = 0 //always visible
	//Behavior_Reset_On_Map_Change Behavior  = 1
	Behavior_Reset_On_Daily_Reset      Behavior = 2
	Behavior_Visible_Before_Activation Behavior = 3 //disapear forever on interact
	Behavior_Reapear_After_Timer       Behavior = 4 //use resetLength as timer
	//Behavior_Reapear_After_Reset Behavior = 5
	Behavior_Once_Per_Instance        Behavior = 6
	Behavior_Once_Daily_Per_Character Behavior = 7
)
